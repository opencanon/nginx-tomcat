/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package transformer

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type objectTransformer struct{}

func NewObjectTransformer() *objectTransformer {
	return &objectTransformer{}
}

func (t *objectTransformer) TransformObjects(namespace string, name string, objects []client.Object) ([]client.Object, error) {

	for i := range objects {
		switch obj := objects[i].(type) {
		// Case 1: Transform Deployments
		case *appsv1.Deployment:
			ensurePodAntiAffinity(obj)       // Add HA rules
			setDefaultResourceLimits(obj)    // Enforce CPU/Memory
			objects[i] = toUnstructured(obj) // Convert back

		// Case 2: Transform Services
		case *corev1.Service:
			if obj.Spec.Type == corev1.ServiceTypeClusterIP {
				obj.Spec.ClusterIP = "None" // Force Headless Service
			}
			objects[i] = toUnstructured(obj)
		}
	}
	return objects, nil
}

func asStatefulSet(object client.Object) *appsv1.StatefulSet {
	if statefulSet, ok := object.(*appsv1.StatefulSet); ok {
		return statefulSet
	}
	if object, ok := object.(*unstructured.Unstructured); ok && (object.GetObjectKind().GroupVersionKind() == schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}) {
		statefulSet := &appsv1.StatefulSet{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, statefulSet); err != nil {
			panic(err)
		}
		return statefulSet
	}
	return nil
}

func asUnstructurable(object client.Object) *unstructured.Unstructured {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		panic(err)
	}
	return &unstructured.Unstructured{Object: m}
}

// ====

// --- Helper Functions ---

// ensurePodAntiAffinity adds anti-affinity rules to Deployments.
func ensurePodAntiAffinity(deploy *appsv1.Deployment) {
	if deploy.Spec.Template.Spec.Affinity == nil {
		deploy.Spec.Template.Spec.Affinity = &corev1.Affinity{}
	}
	if deploy.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
		deploy.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
	}

	// Prefer spreading pods across hosts
	deploy.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
		{
			Weight: 100,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey:   "kubernetes.io/hostname",
				LabelSelector: deploy.Spec.Selector,
			},
		},
	}
}

// setDefaultResourceLimits enforces CPU/Memory limits.
func setDefaultResourceLimits(deploy *appsv1.Deployment) {
	for i := range deploy.Spec.Template.Spec.Containers {
		container := &deploy.Spec.Template.Spec.Containers[i]
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}
		// Set defaults if not specified
		if _, exists := container.Resources.Limits[corev1.ResourceCPU]; !exists {
			container.Resources.Limits[corev1.ResourceCPU] = resource.MustParse("123m")
		}
		if _, exists := container.Resources.Limits[corev1.ResourceMemory]; !exists {
			container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("122Mi")
		}
	}
}

// toUnstructured converts a typed object to unstructured.Unstructured.
func toUnstructured(obj client.Object) *unstructured.Unstructured {
	unstructuredObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return &unstructured.Unstructured{Object: unstructuredObj}
}
