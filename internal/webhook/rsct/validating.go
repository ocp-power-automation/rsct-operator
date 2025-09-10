package rsct

import (
	"context"
	"fmt"

	rsctv1alpha1 "github.com/ocp-power-automation/rsct-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type RSCTValidator struct {
	Client client.Client
}

// ValidateCreate implements admission.CustomValidator.
func (r *RSCTValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr, ok := obj.(*rsctv1alpha1.RSCT)
	if !ok {
		return nil, fmt.Errorf("expected RSCT, got %T", obj)
	}

	var crList rsctv1alpha1.RSCTList
	if err := r.Client.List(ctx, &crList); err != nil {
		return nil, fmt.Errorf("cannot list RSCT: %w", err)
	}

	if len(crList.Items) > 0 {
		return nil, fmt.Errorf("only one RSCT instance is allowed (found %d), rejecting creation of %s", len(crList.Items), cr.Name)
	}

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator.
func (r *RSCTValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator.
func (r *RSCTValidator) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

var _ admission.CustomValidator = &RSCTValidator{}

// Register the webhook with the manager
func RegisterWebhooks(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&rsctv1alpha1.RSCT{}).
		WithValidator(&RSCTValidator{Client: mgr.GetClient()}).
		Complete()
}
