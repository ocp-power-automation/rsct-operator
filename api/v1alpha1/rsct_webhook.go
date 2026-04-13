/*
Copyright 2024.

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

package v1alpha1

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var rsctlog = logf.Log.WithName("rsct-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *RSCT) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &RSCT{}).
		WithValidator(&RSCTValidator{Client: mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/validate-rsct-ibm-com-v1alpha1-rsct,mutating=false,failurePolicy=fail,sideEffects=None,groups=rsct.ibm.com,resources=rscts,verbs=create;update,versions=v1alpha1,name=vrsct.kb.io,admissionReviewVersions=v1

// +kubebuilder:object:generate=false
// RSCTValidator validates RSCT resources
type RSCTValidator struct {
	Client client.Client
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *RSCTValidator) ValidateCreate(ctx context.Context, obj *RSCT) (admission.Warnings, error) {
	rsctlog.Info("validate create", "name", obj.Name)

	rsctList := &RSCTList{}
	if err := v.Client.List(ctx, rsctList); err != nil {
		return nil, fmt.Errorf("failed to list RSCT instances: %w", err)
	}

	if len(rsctList.Items) > 0 {
		return nil, fmt.Errorf("an RSCT instance already exists in the cluster; only one instance is allowed")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *RSCTValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *RSCT) (admission.Warnings, error) {
	rsctlog.Info("validate update", "name", newObj.Name)

	// Update is allowed for the existing instance
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *RSCTValidator) ValidateDelete(ctx context.Context, obj *RSCT) (admission.Warnings, error) {
	rsctlog.Info("validate delete", "name", obj.Name)

	// Delete is always allowed
	return nil, nil
}
