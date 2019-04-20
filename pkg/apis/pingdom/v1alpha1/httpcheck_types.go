/*
Copyright 2019 Fabian Sabau <fabian.sabau@gmail.com>.

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

type PingdomStatus string

var (
	StatusSuccess PingdomStatus = "Succeeded"
	StatusFail    PingdomStatus = "Fail"
)

// HttpCheckSpec defines the desired state of HttpCheck
type HttpCheckSpec struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// HttpCheckStatus defines the observed state of HttpCheck
type HttpCheckStatus struct {
	PingdomID     int           `json:"pingdomId,omitempty"`
	PingdomStatus PingdomStatus `json:"pingdomStatus,omitempty"`
	Error         string        `json:"error,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HttpCheck is the Schema for the httpchecks API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type HttpCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HttpCheckSpec   `json:"spec,omitempty"`
	Status HttpCheckStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HttpCheckList contains a list of HttpCheck
type HttpCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HttpCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HttpCheck{}, &HttpCheckList{})
}
