package kubernetes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/go-common/kubernetes/errors"
)

// DoUntilNotConflict always executes the (kubernetes object update) function until there is no conflict.
func DoUntilNotConflict(f func() error) error {
	err := f()
	for errors.IsConflict(err) {
		err = f()
	}
	return err
}

// DoUntilNotConflictWithin is the same as DoUntilNotConflict, but can limit the max retry count.
func DoUntilNotConflictWithin(f func() error, count int) error {
	if count < 0 {
		return fmt.Errorf("invalid count %v", count)
	}
	var err error
	for count >= 0 {
		err = f()
		if err == nil {
			return nil
		}

		if errors.IsConflict(err) {
			count--
			continue
		}
		return err
	}
	return fmt.Errorf("exceeded retry count with error: %v", err)
}

// IgnoredNamespace returns true if the given namespace belongs to one of the
// following cases:
//   kube-system namespace
//   kube-public namespace
//   default namespace
//   kube-node-lease namespace
func IgnoredNamespace(ns string) bool {
	return ns == metav1.NamespaceSystem ||
		ns == metav1.NamespacePublic ||
		ns == metav1.NamespaceDefault ||
		ns == corev1.NamespaceNodeLease
}
