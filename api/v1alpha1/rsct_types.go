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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RSCTSpec defines the desired state of RSCT
type RSCTSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image is an RSCT image
	// +kubebuilder:default="quay.io/powercloud/rsct-ppc64le:latest"
	// +optional
	Image *string `json:"image,omitempty"`
}

// RSCTStatus defines the observed state of RSCT
type RSCTStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State *string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RSCT is the Schema for the rscts API
type RSCT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RSCTSpec   `json:"spec,omitempty"`
	Status RSCTStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RSCTList contains a list of RSCT
type RSCTList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RSCT `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RSCT{}, &RSCTList{})
}
