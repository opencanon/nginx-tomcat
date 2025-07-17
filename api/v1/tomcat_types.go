/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/sap/component-operator-runtime/pkg/component"
	componentoperatorruntimetypes "github.com/sap/component-operator-runtime/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TomcatSpec defines the desired state of Tomcat.
type TomcatSpec struct {
	Version string `json:"version,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas int `json:"replicas,omitempty"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Resources TomcatResources `json:"resources,omitempty"`
}

type TomcatResources struct {
	Requests TomcatRequests `json:"requests,omitempty"`
}
type TomcatRequests struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// TomcatStatus defines the observed state of Tomcat.
type TomcatStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	component.Status `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Tomcat is the Schema for the tomcats API.
type Tomcat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TomcatSpec   `json:"spec,omitempty"`
	Status TomcatStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TomcatList contains a list of Tomcat.
type TomcatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tomcat `json:"items"`
}

func (c *Tomcat) GetSpec() componentoperatorruntimetypes.Unstructurable {
	return &c.Spec
}

func init() {
	SchemeBuilder.Register(&Tomcat{}, &TomcatList{})
}

func (s *TomcatSpec) ToUnstructured() map[string]any {
	result, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
	if err != nil {
		panic(err)
	}
	if namespace, ok := result["namespace"]; ok {
		if _, ok := namespace.(string); !ok {
			panic("spec.namespace set but is not a string")
		}
	}
	if name, ok := result["name"]; ok {
		if _, ok := name.(string); !ok {
			panic("spec.name set but is not a string")
		}
	}
	return result
}

func (c *Tomcat) GetStatus() *component.Status {
	return &c.Status.Status
}
