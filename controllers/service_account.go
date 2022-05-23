/*
Copyright 2021.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1alpha1 "github.com/mjturek/rsct-operator/api/v1alpha1"
)

// ensureExternalDNSServiceAccount ensures that the externalDNS service account exists.
func (r *RSCTReconciler) ensureRSCTServiceAccount(ctx context.Context, namespace string, rsct *operatorv1alpha1.RSCT) (bool, *corev1.ServiceAccount, error) {

	// TODO(mjturek): Do this in a less hardcoded fashion.
	nsName := types.NamespacedName{Namespace: "powervm-rmc", Name: "powervm-rmc"}

	desired := desiredRSCTServiceAccount(nsName)

	if err := controllerutil.SetControllerReference(rsct, desired, r.Scheme); err != nil {
		return false, nil, fmt.Errorf("failed to set the controller reference for service account: %w", err)
	}

	exist, current, err := r.currentRSCTServiceAccount(ctx, nsName)
	if err != nil {
		return false, nil, err
	}

	if !exist {
		if err := r.createRSCTServiceAccount(ctx, desired); err != nil {
			return false, nil, err
		}
		return r.currentRSCTServiceAccount(ctx, nsName)
	}

	return true, current, nil
}

// currentExternalDNSServiceAccount gets the current externalDNS service account resource.
func (r *RSCTReconciler) currentRSCTServiceAccount(ctx context.Context, nsName types.NamespacedName) (bool, *corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, nsName, sa); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, sa, nil
}

// desiredExternalDNSServiceAccount returns the desired serivce account resource.
func desiredRSCTServiceAccount(nsName types.NamespacedName) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
	}
}

// createRSCTServiceAccount creates the given service account using the reconciler's client.
func (r *RSCTReconciler) createRSCTServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
	if err := r.Client.Create(ctx, sa); err != nil {
		return fmt.Errorf("failed to create RSCT service account %s/%s: %w", sa.Namespace, sa.Name, err)
	}

	r.log.Info("created RSCT service account", "namespace", sa.Namespace, "name", sa.Name)
	return nil
}
