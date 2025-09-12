package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SecretWatcherSpec defines the desired state of SecretWatcher
type SecretWatcherSpec struct {
	SecretName string `json:"secretName"`
}

// SecretWatcherStatus defines the observed state of SecretWatcher
type SecretWatcherStatus struct {
	LastRotated metav1.Time `json:"lastRotated,omitempty"`
}

// +kubebuilder:object:root=true

type SecretWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretWatcherSpec   `json:"spec,omitempty"`
	Status SecretWatcherStatus `json:"status,omitempty"`
}

func (in *SecretWatcher) DeepCopyObject() runtime.Object {
	return in
}

// +kubebuilder:object:root=true

type SecretWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretWatcher `json:"items"`
}

func (in *SecretWatcherList) DeepCopyObject() runtime.Object {
	return in
}

// Register with scheme
func AddToScheme(s *runtime.Scheme) error {
	s.AddKnownTypes(schema.GroupVersion{Group: "example.com", Version: "v1"},
		&SecretWatcher{},
		&SecretWatcherList{},
	)
	metav1.AddToGroupVersion(s, schema.GroupVersion{Group: "example.com", Version: "v1"})
	return nil
}
