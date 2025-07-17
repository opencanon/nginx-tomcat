/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package operator

import (
	mytomcat "opencanon.com/api/v1"
	_ "opencanon.com/internal/generator"

	"embed"
	"flag"

	"github.com/pkg/errors"
	"github.com/sap/component-operator-runtime/pkg/component"

	"github.com/sap/component-operator-runtime/pkg/manifests"
	"github.com/sap/component-operator-runtime/pkg/manifests/helm"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sap/component-operator-runtime/pkg/operator"
	"opencanon.com/internal/transformer"
)

const Name = "tomcat.seneca"

//go:embed all:data
var data embed.FS

type Options struct {
	Name                  string
	DefaultServiceAccount string
	FlagPrefix            string
}

type Operator struct {
	options Options
}

var defaultOperator operator.Operator = New()

func GetName() string {
	return defaultOperator.GetName()
}

func InitScheme(scheme *runtime.Scheme) {
	defaultOperator.InitScheme(scheme)
}

func InitFlags(flagset *flag.FlagSet) {
	defaultOperator.InitFlags(flagset)
}

func ValidateFlags() error {
	return defaultOperator.ValidateFlags()
}

func GetUncacheableTypes() []client.Object {
	return defaultOperator.GetUncacheableTypes()
}

func Setup(mgr ctrl.Manager) error {
	return defaultOperator.Setup(mgr)
}

func New() *Operator {
	return NewWithOptions(Options{})
}

func NewWithOptions(options Options) *Operator {
	operator := &Operator{options: options}
	if operator.options.Name == "" {
		operator.options.Name = Name
	}
	return operator
}

func (o *Operator) GetName() string {
	return o.options.Name
}

func (o *Operator) InitScheme(scheme *runtime.Scheme) {
	utilruntime.Must(mytomcat.AddToScheme(scheme))
}

func (o *Operator) InitFlags(flagset *flag.FlagSet) {

	flagset.StringVar(&o.options.DefaultServiceAccount,
		"default-service-account",
		o.options.DefaultServiceAccount,
		"Default service account name")
}

func (o *Operator) ValidateFlags() error {
	return nil
}

func (o *Operator) GetUncacheableTypes() []client.Object {
	return []client.Object{&mytomcat.Tomcat{}}
}

func (o *Operator) Setup(mgr ctrl.Manager) error {

	parameterTransformer, err := manifests.NewTemplateParameterTransformer(data, "data/parameters.yaml")
	if err != nil {
		return errors.Wrap(err, "error initializing parameter transformer")
	}
	objectTransformer := transformer.NewObjectTransformer()
	resourceGenerator, err := helm.NewTransformableHelmGenerator(
		data,
		"data/charts/tomcat",
		mgr.GetClient(),
	)

	if err != nil {
		return errors.Wrap(err, "====error initializing resource generator")
	}
	resourceGenerator.
		WithParameterTransformer(parameterTransformer).
		WithObjectTransformer(objectTransformer)

	// TODO: handle increases of persistence.size somehow (instead of making it immutable)
	// this would require to recreate the statefulset (since persistentVolumeClaimTemplate is immutable)
	// and to extend existing persistent volume claims (supposing that they are resizable)

	if err := component.NewReconciler[*mytomcat.Tomcat](
		o.options.Name,
		resourceGenerator,
		component.ReconcilerOptions{},
	).WithPostReconcileHook(
		resourcesRequesetBinding,
	).WithPostReconcileHook(
		reconcileBinding,
	).SetupWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create controller")
	}

	mytomcat.NewWebhook().SetupWithManager(mgr)

	return nil

}
