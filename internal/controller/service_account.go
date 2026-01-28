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

	operatorv1alpha1 "github.com/ocp-power-automation/rsct-operator/api/v1alpha1"
	securityv1 "github.com/openshift/api/security/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ensureRSCTServiceAccount ensures that the RSCT service account exists.
func (r *RSCTReconciler) ensureRSCTServiceAccount(ctx context.Context, rsct *operatorv1alpha1.RSCT) (bool, *corev1.ServiceAccount, error) {
	nsName := types.NamespacedName{Namespace: rsct.Namespace, Name: rsct.Name}

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

// currentRSCTServiceAccount gets the current RSCT service account resource and ensures it has privileged SCC.
func (r *RSCTReconciler) currentRSCTServiceAccount(ctx context.Context, nsName types.NamespacedName) (bool, *corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, nsName, sa); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}

	// if non-OpenShift cluster, skip SCC logic
	gvk := schema.GroupVersionKind{
		Group:   "security.openshift.io",
		Version: "v1",
		Kind:    "SecurityContextConstraints",
	}

	mapper := r.Client.RESTMapper()
	_, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if meta.IsNoMatchError(err) {
		// skip SCC logic
		return true, sa, nil
	} else if err != nil {
		return true, sa, fmt.Errorf("failed to check SCC API: %w", err)
	}

	// for OpenShift, ensure the service account has privileged SCC
	scc := &securityv1.SecurityContextConstraints{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: "privileged"}, scc); err != nil {
		return true, sa, fmt.Errorf("error getting privileged SCC: %w", err)
	}

	saUser := fmt.Sprintf("system:serviceaccount:%s:%s", nsName.Namespace, nsName.Name)

	if !contains(scc.Users, saUser) {
		patch := client.MergeFrom(scc.DeepCopy())
		scc.Users = append(scc.Users, saUser)

		if err := r.Client.Patch(ctx, scc, patch); err != nil {
			return true, sa, fmt.Errorf("failed to patch privileged SCC: %w", err)
		}
	}

	return true, sa, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// desiredRSCTServiceAccount returns the desired serivce account resource.
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

	return nil
}
