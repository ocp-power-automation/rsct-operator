/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rsctv1alpha1 "github.com/mjturek/rsct-operator/api/v1alpha1"
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
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// the RSCT object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RSCTReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rsct := &rsctv1alpha1.RSCT{}
	if err := r.Client.Get(ctx, req.NamespacedName, rsct); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get RSCT %s: %w", req, err)
	}

	// TODO(mjturek): Allow image specification
	r.Config.Namespace = rsct.Namespace
	r.Config.Name = rsct.Name
	r.Config.Image = "quay.io/powercloud/rsct-ppc64le:latest"

	haveServiceAccount, sa, err := r.ensureRSCTServiceAccount(ctx, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure powervm-rmc service account: %w", err)
	} else if !haveServiceAccount {
		return reconcile.Result{}, fmt.Errorf("failed to get powervm-rmc service account: %w", err)
	}

	_, currentDaemonSet, err := r.ensureRSCTDaemonSet(ctx, sa, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure powervm-rmc daemonSet: %w", err)
	}
	if err := r.updateRSCTStatus(ctx, rsct, currentDaemonSet); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to update RSCT custom resource %s: %w", rsct.Name, err)
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RSCTReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rsctv1alpha1.RSCT{}).
		Complete(r)
}
