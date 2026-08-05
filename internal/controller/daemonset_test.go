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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
