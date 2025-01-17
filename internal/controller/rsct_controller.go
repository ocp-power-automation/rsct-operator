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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rsctv1alpha1 "github.com/ocp-power-automation/rsct-operator/api/v1alpha1"
)

type RSCTConfig struct {
	// Namespace is the namespace that RSCT should be deployed in.
	Namespace string
	// Name is the name of the operand
	Name string
	// Image is the RSCT image to use.
	Image string
}

// RSCTReconciler reconciles a RSCT object
type RSCTReconciler struct {
	Config RSCTConfig
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RSCT object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RSCTReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	rsct := &rsctv1alpha1.RSCT{}
	if err := r.Client.Get(ctx, req.NamespacedName, rsct); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get RSCT %s: %w", req, err)
	}

	r.Config.Namespace = rsct.Namespace
	r.Config.Name = rsct.Name
	// Set default RSCT image if not specified
	if rsct.Spec.Image != nil {
		r.Config.Image = *rsct.Spec.Image
	}

	haveServiceAccount, sa, err := r.ensureRSCTServiceAccount(ctx, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure rsct service account: %w", err)
	} else if !haveServiceAccount {
		return reconcile.Result{}, fmt.Errorf("failed to get rsct service account: %w", err)
	}

	_, currentDaemonSet, err := r.ensureRSCTDaemonSet(ctx, sa, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure rsct daemonSet: %w", err)
	}
	if err := r.updateRSCTStatus(ctx, rsct, currentDaemonSet); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to update RSCT custom resource %s: %w", rsct.Name, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RSCTReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rsctv1alpha1.RSCT{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
