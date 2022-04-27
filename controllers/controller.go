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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rsctv1alpha1 "github.com/mjturek/rsct-operator/api/v1alpha1"
)

type Config struct {
	Name string
	// Namespace is the namespace that RSCT should be deployed in.
	Namespace string
	// Image is the RSCT image to use.
	Image string
}

// RSCTReconciler reconciles a RSCT object
type RSCTReconciler struct {
	config Config
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rsct.ibm.com,resources=rscts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RSCT object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RSCTReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.log.WithValues("RSCT", req.NamespacedName)
	reqLogger.Info("reconciling RSCT")

	rsct := &rsctv1alpha1.RSCT{}
	if err := r.client.Get(ctx, req.NamespacedName, rsct); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("RSCT not found; reconciliation will be skipped")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get RSCT %s: %w", req, err)
	}

	haveServiceAccount, sa, err := r.ensureRSCTServiceAccount(ctx, r.config.Namespace, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure powervm-rmc service account: %w", err)
	} else if !haveServiceAccount {
		return reconcile.Result{}, fmt.Errorf("failed to get powervm-rmc service account: %w", err)
	}

	_, currentDaemonSet, err := r.ensureRSCTDaemonSet(ctx, r.config.Namespace, r.config.Image, sa, rsct)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to ensure powervm-rmc daemonSet: %w", err)
	}
	if err := r.updateRSCTStatus(ctx, rsct, currentDaemonSet); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to update RSCT custom resource %s: %w", RSCT.Name, err)
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RSCTReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rsctv1alpha1.RSCT{}).
		Complete(r)
}
