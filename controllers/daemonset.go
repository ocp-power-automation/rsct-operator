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

	"github.com/google/go-cmp/cmp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	rsctBaseName        = "powervm-rmc"
	arch                = "ppc64le"
	masterNodeRoleLabel = "node-role.kubernetes.io/master"
	osID                = "rhcos"
	rmcPort             = 657
	rmcAppName          = "powervm-rmc"
)

// ensureRSCTDaemonSet ensures that the RSCT DaemonSet xists.
// Returns a Boolean value indicating whether the daemonSet exists, a pointer to the daemonSet, and an error when relevant.
func (r *RSCTReconciler) ensureRSCTDaemonSet(ctx context.Context, namespace, image string, serviceAccount *corev1.ServiceAccount, rsct *rsctv1alpha1.RSCT) (bool, *appsv1.Deployment, error) {

	// TODO(mjturek): Do this in a less hardcoded fashion.
	nsName := types.NamespacedName{Namespace: "powervm-rmc", Name: "powervm-rmc"}

	desired, err := desiredRSCTDaemonSet()
	if err != nil {
		return false, nil, fmt.Errorf("failed to build RSCT daemonSet: %w", err)
	}

	if err := controllerutil.SetControllerReference(rsct, desired, r.scheme); err != nil {
		return false, nil, fmt.Errorf("failed to set the controller reference for daemonSet: %w", err)
	}

	exist, current, err := r.currentRSCTDaemonSet(ctx, nsName)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current RSCT daemonSet: %w", err)
	}

	// create the deployment
	if !exist {
		if err := r.createRSCTDaemonSet(ctx, desired); err != nil {
			return false, nil, err
		}
		return r.currentRSCTDaemonSet(ctx, nsName)
	}

	// update the deployment
	if updated, err := r.updateRSCT(ctx, current, desired); err != nil {
		return true, current, err
	} else if updated {
		return r.currentRSCTDaemonSet(ctx, nsName)
	}

	return true, current, nil
}

// currentExternalDNSDeployment gets the current externalDNS deployment resource.
func (r *RSCTReconciler) currentRSCTDaemonSet(ctx context.Context, nsName types.NamespacedName) (bool, *appsv1.Deployment, error) {
	depl := &appsv1.Deployment{}
	if err := r.client.Get(ctx, nsName, depl); err != nil {
		if errors.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, depl, nil
}

// desiredExternalDNSDeployment returns the desired deployment resource.
func desiredRSCTDaemonSet(config *RSCTConfig) (*appsv1.DaemonSet, error) {

	matchLabels := map[string]string{
		"app": "powervm-rmc",
	}

	nodeSelectorLabels := map[string]string{
		"kubernetes.io/arch":      "ppc64le",
		"node.openshift.io/os_id": "rhcos",
	}

	tolerations := []corev1.Toleration{
		{
			Key:      masterNodeRoleLabel,
			Operator: corev1.TolerationOpExists,
		},
	}

	volumes := []corev1.Volume{
		{
			Name: "lib-modules",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/lib/modules",
				},
			},
		},
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  config.Name,
							Image: config.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: rmcPort,
									HostPort:      rmcPort,
									Name:          "rmc-tcp",
									Protocol:      corev1.ProtocolTCP,
								},
								{
									ContainerPort: rmcPort,
									HostPort:      rmcPort,
									Name:          "rmc-udp",
									Protocol:      corev1.ProtocolUDP,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: config.memoryLimit,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    config.cpuRequest,
									corev1.ResourceMemory: config.memoryRequest,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "lib-modules",
									MountPath: "/lib/modules",
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: true,
								RunAsUser:  0,
							},
						},
					},
					HostNetwork:        true,
					NodeSelector:       nodeSelectorLabels,
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: config.Name,
					Volumes:            volumes,
					Tolerations:        tolerations,
				},
			},
		},
	}

	return ds, nil
}

// createExternalDNSDeployment creates the given deployment using the reconciler's client.
func (r *RSCTReconciler) createRSCTDaemonSet(ctx context.Context, ds *appsv1.DaemonSet) error {
	if err := r.client.Create(ctx, ds); err != nil {
		return fmt.Errorf("failed to create RSCT daemonset %s/%s: %w", ds.Namespace, ds.Name, err)
	}
	r.log.Info("created RSCT daemonset", "namespace", ds.Namespace, "name", ds.Name)
	return nil
}

// updateRSCTDaemonSet updates the in-cluster RSCT daemonset.
// Returns a boolean indicating if an update was made, and an error when relevant.
func (r *RSCTReconciler) updateRSCTDaemonSet(ctx context.Context, current, desired *appsv1.DaemonSet) (bool, error) {
	changed, updated := rsctDaemonSetChanged(current, desired)
	if !changed {
		return false, nil
	}

	if err := r.client.Update(ctx, updated); err != nil {
		return false, fmt.Errorf("failed to update RSCT DaemonSet %s/%s: %w", desired.Namespace, desired.Name, err)
	}
	r.log.Info("updated RSCT DaemonSet", "namespace", desired.Namespace, "name", desired.Name)
	return true, nil
}

// rsctDaemonSetChanged returns a boolean indicating if an update is needed and the desired daemonset.
func rsctDaemonSetChanged(current, expected *appsv1.DaemonSet) (bool, *appsv1.DaemonSet) {
	updated := current.DeepCopy()
	//TODO(mjturek): Do what the comment says
	return true, nil
}
