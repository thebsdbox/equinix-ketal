/*
Copyright 2021 Dan Finneran.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EipSpec defines the desired state of Eip
type EipSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// UUID is the unique identifier for an Equinix Metal EIP
	UUID string `json:"uuid"`

	// Address is the external address of the EIP
	Address string `json:"address"`

	// Public defines if this is an internal or external EIP
	Public bool `json:"public"`
}

// EipStatus defines the observed state of Eip
type EipStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Eip is the Schema for the eips API
type Eip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EipSpec   `json:"spec,omitempty"`
	Status EipStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EipList contains a list of Eip
type EipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Eip `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Eip{}, &EipList{})
}
