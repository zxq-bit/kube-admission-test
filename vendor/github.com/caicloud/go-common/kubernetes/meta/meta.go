package meta

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/go-common/slice"
)

// GetLabel returns the value of the given key from labels.
func GetLabel(obj metav1.Object, key string) string {
	if interfaces.IsNil(obj) {
		return ""
	}
	labels := obj.GetLabels()
	if labels == nil {
		return ""
	}
	return labels[key]
}

// SetLabel sets or updates a label for the given object, returns true if the label updates.
func SetLabel(obj metav1.Object, key, value string) bool {
	if interfaces.IsNil(obj) {
		return false
	}
	var updated bool
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 1)
		updated = true
	} else {
		prev, ok := labels[key]
		updated = !ok || prev != value
	}
	labels[key] = value
	obj.SetLabels(labels)
	return updated
}

// GetAnnotation returns the value of the given key from annotations.
func GetAnnotation(obj metav1.Object, key string) string {
	if interfaces.IsNil(obj) {
		return ""
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return ""
	}
	return annotations[key]
}

// SetAnnotation sets or updates an annotation for the given object, returns true if the annotation updates.
func SetAnnotation(obj metav1.Object, key, value string) bool {
	if interfaces.IsNil(obj) {
		return false
	}
	var updated bool
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string, 1)
		updated = true
	} else {
		prev, ok := annotations[key]
		updated = !ok || prev != value
	}
	annotations[key] = value
	obj.SetAnnotations(annotations)
	return updated
}

// AddFinalizer adds a new finalizer to the object, returns true if the Finalizers updates.
func AddFinalizer(obj metav1.Object, finalizer string) bool {
	if interfaces.IsNil(obj) {
		return false
	}
	finalizers := obj.GetFinalizers()
	if slice.StringInSlice(finalizers, finalizer) {
		return false
	}
	obj.SetFinalizers(append(finalizers, finalizer))
	return true
}

// RemoveFinalizer removes the finalizer from the object, returns true if the Finalizers updates.
func RemoveFinalizer(obj metav1.Object, finalizer string) bool {
	if interfaces.IsNil(obj) {
		return false
	}
	finalizers := obj.GetFinalizers()
	if !slice.StringInSlice(finalizers, finalizer) {
		return false
	}
	obj.SetFinalizers(slice.RemoveStringInSlice(finalizers, finalizer))
	return true
}
