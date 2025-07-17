/*
SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

package v1

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:object:generate=false
type Webhook struct {
}

// var _ admission.ValidatingWebhook[*Redis] = &Webhook{}

func NewWebhook() *Webhook {
	return &Webhook{}
}

func (w *Webhook) SetupWithManager(mgr manager.Manager) {

}
