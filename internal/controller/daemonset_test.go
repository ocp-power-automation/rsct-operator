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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func serviceAccount(name, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// TestDesiredRSCTDaemonSet_ContainerNameIsAlwaysRmcAppName verifies that the
// container name is always hardcoded to rmcAppName regardless of the CR name
// passed via DaemonSetConfig.Name.
func TestDesiredRSCTDaemonSet_ContainerNameIsAlwaysRmcAppName(t *testing.T) {
	crNames := []string{
		"rsct",
		"my-rsct",
		"rsct-3.3.3.6-ubi9-test",
		"rsct.with.dots",
		"x",
	}

	for _, crName := range crNames {
		t.Run(crName, func(t *testing.T) {
			config := &DaemonSetConfig{
				Namespace:      "default",
				Name:           crName,
				Image:          "example.com/rsct:latest",
				MemoryLimit:    "1Gi",
				MemoryRequest:  "500Mi",
				CPURequest:     "0.1",
				ServiceAccount: serviceAccount(crName, "default"),
			}

			ds, err := desiredRSCTDaemonSet(config)
			if err != nil {
				t.Fatalf("desiredRSCTDaemonSet(%q) returned unexpected error: %v", crName, err)
			}

			containers := ds.Spec.Template.Spec.Containers
			if len(containers) != 1 {
				t.Fatalf("expected 1 container, got %d", len(containers))
			}

			if got := containers[0].Name; got != rmcAppName {
				t.Errorf("container name = %q, want %q (rmcAppName); CR name was %q",
					got, rmcAppName, crName)
			}
		})
	}
}

// TestReconcileRSCTDaemonSetImage verifies reconcileRSCTDaemonSetImage behaviour
// across the three cases: image changed, image unchanged, and no containers (guard).
func TestReconcileRSCTDaemonSetImage(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add appsv1 to scheme: %v", err)
	}

	makeDS := func(image string) *appsv1.DaemonSet {
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rsct",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: rmcAppName, Image: image},
						},
					},
				},
			},
		}
	}

	t.Run("image changed", func(t *testing.T) {
		current := makeDS("quay.io/powercloud/rsct-ppc64le:v1")
		desired := makeDS("quay.io/powercloud/rsct-ppc64le:v2")

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build()
		r := &RSCTReconciler{Client: fakeClient, Scheme: scheme}

		returned, err := r.reconcileRSCTDaemonSetImage(context.Background(), current, desired)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The returned object must already carry the new image.
		if got := returned.Spec.Template.Spec.Containers[0].Image; got != desired.Spec.Template.Spec.Containers[0].Image {
			t.Errorf("returned image = %q, want %q", got, desired.Spec.Template.Spec.Containers[0].Image)
		}

		// The persisted object must also carry the new image.
		stored := &appsv1.DaemonSet{}
		if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(current), stored); err != nil {
			t.Fatalf("failed to get DaemonSet after update: %v", err)
		}
		if got := stored.Spec.Template.Spec.Containers[0].Image; got != desired.Spec.Template.Spec.Containers[0].Image {
			t.Errorf("stored image = %q, want %q", got, desired.Spec.Template.Spec.Containers[0].Image)
		}
	})

	t.Run("image unchanged", func(t *testing.T) {
		const img = "quay.io/powercloud/rsct-ppc64le:latest"
		current := makeDS(img)
		desired := makeDS(img)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build()
		r := &RSCTReconciler{Client: fakeClient, Scheme: scheme}

		returned, err := r.reconcileRSCTDaemonSetImage(context.Background(), current, desired)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if returned != current {
			t.Error("expected the original current pointer to be returned unchanged")
		}

		// Confirm no write occurred — resourceVersion should be unchanged.
		stored := &appsv1.DaemonSet{}
		if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(current), stored); err != nil {
			t.Fatalf("failed to get DaemonSet: %v", err)
		}
		if stored.ResourceVersion != current.ResourceVersion {
			t.Errorf("resourceVersion changed from %q to %q; unexpected write occurred",
				current.ResourceVersion, stored.ResourceVersion)
		}
	})

	t.Run("no containers returns error", func(t *testing.T) {
		current := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "rsct", Namespace: "default"},
		}
		desired := makeDS("quay.io/powercloud/rsct-ppc64le:v2")

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build()
		r := &RSCTReconciler{Client: fakeClient, Scheme: scheme}

		if _, err := r.reconcileRSCTDaemonSetImage(context.Background(), current, desired); err == nil {
			t.Error("expected an error for a DaemonSet with no containers, got nil")
		}
	})
}
