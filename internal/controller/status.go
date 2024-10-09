package controller

import (
	"context"
	"fmt"
	"slices"

	rsctv1alpha1 "github.com/ocp-power-automation/rsct-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// matchPodsStatus checks if status of all the pods in slice are same or not
func matchPodsStatus(podsStatus []string, status string) bool {
	for _, ps := range podsStatus {
		if ps == status {
			continue
		}
		return false
	}
	return true
}

// evalOperatorStatus determines operator status based on the pods status
func evalOperatorStatus(podList *corev1.PodList) string {
	var effectiveStatus string
	var podsStatus []string
	for _, pod := range podList.Items {
		switch {
		case pod.Status.Phase == corev1.PodPending:
			podsStatus = append(podsStatus, "PENDING")
		case pod.Status.Phase == corev1.PodFailed:
			podsStatus = append(podsStatus, "FAILED")
		case pod.Status.Phase == corev1.PodRunning:
			podsStatus = append(podsStatus, "RUNNING")
			continue
		default:
			podsStatus = append(podsStatus, "UNKNOWN")
		}
	}

	if slices.Contains(podsStatus, "RUNNING") {
		if slices.Contains(podsStatus, "FAILED") || slices.Contains(podsStatus, "PENDING") || slices.Contains(podsStatus, "UNKNOWN") {
			effectiveStatus = "PARTIALLY_RUNNING"
		} else {
			effectiveStatus = "RUNNING"
		}
	} else if matchPodsStatus(podsStatus, "FAILED") {
		effectiveStatus = "FAILED"
	} else if matchPodsStatus(podsStatus, "PENDING") {
		effectiveStatus = "PENDING"
	}
	return effectiveStatus
}

// updateRSCTStatus updates RSCT operator status
func (r *RSCTReconciler) updateRSCTStatus(ctx context.Context, rsct *rsctv1alpha1.RSCT, currentDaemonSet *appsv1.DaemonSet) error {
	// Operator status:
	// 1. PENDING
	// 2. RUNNING
	// 3. PARTIALLY-RUNNING
	// 4. FAILED

	pods := &corev1.PodList{}

	labelSelector := labels.SelectorFromSet(map[string]string{"app": currentDaemonSet.Name})
	listOpts := &client.ListOptions{Namespace: rsct.Namespace, LabelSelector: labelSelector}
	listOpts.ApplyOptions([]client.ListOption{})

	if err := r.List(ctx, pods, listOpts); err != nil {
		return fmt.Errorf("failed to get list of rsct operator pods: %w", err)
	}

	operatorStatus := evalOperatorStatus(pods)
	rsct.Status.CurrentStatus = &operatorStatus

	err := r.Status().Update(ctx, rsct)
	if err != nil {
		return err
	}

	return nil
}
