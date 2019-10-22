package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExamplekindSpec defines the desired state of Examplekind
type ArctiqOperSpec struct {
	Count int32  `json:"count"`
	Group string `json:"group"`
	Image string `json:"image"`
	Port  int32  `json:"port"`
}

// ExamplekindStatus defines the observed state of Examplekind
type ArctiqOperStatus struct {
	PodNames []string `json:"podnames"`
	AppGroup string   `json:"appgroup"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArctiqOper is the Schema for the arctiqopers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ArctiqOper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArctiqOperSpec   `json:"spec,omitempty"`
	Status ArctiqOperStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArctiqOperList contains a list of ArctiqOper
type ArctiqOperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArctiqOper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArctiqOper{}, &ArctiqOperList{})
}
