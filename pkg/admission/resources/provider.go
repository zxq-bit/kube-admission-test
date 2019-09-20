package resources

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceMutateAndValidate mutate and validate
type ResourceMutateAndValidate interface {
	Mutate(interface{}, []byte) ([]byte, bool, error)
	Validate(*v1beta1.AdmissionRequest) (interface{}, interface{}, error)
}

// ResourceProvider provider resource Mutate and Validate
var ResourceProvider map[metav1.GroupVersionResource]ResourceMutateAndValidate

// GetResourceMutateAndValidation gets resource
func GetResourceMutateAndValidation(r metav1.GroupVersionResource) (ResourceMutateAndValidate, bool) {
	rmv, ok := ResourceProvider[r]
	return rmv, ok
}

// Register ...
func Register(gvr metav1.GroupVersionResource, rmv ResourceMutateAndValidate) {
	if ResourceProvider == nil {
		ResourceProvider = make(map[metav1.GroupVersionResource]ResourceMutateAndValidate)
	}
	ResourceProvider[gvr] = rmv
}

// GetGVRList gets resgiter gvr info
func GetGVRList() []metav1.GroupVersionResource {
	list := make([]metav1.GroupVersionResource, 0, len(ResourceProvider))
	for k := range ResourceProvider {
		list = append(list, k)
	}
	return list
}
