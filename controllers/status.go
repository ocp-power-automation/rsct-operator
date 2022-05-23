package controllers

import (
	"context"
	rsctv1alpha1 "github.com/mjturek/rsct-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// updateRSCTStatus will do something magical one day.
func (r *RSCTReconciler) updateRSCTStatus(ctx context.Context, rsct *rsctv1alpha1.RSCT, currentDaemonSet *appsv1.DaemonSet) error {
	return nil
}
