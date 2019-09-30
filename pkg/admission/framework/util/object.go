package util

import (
	"path"
	"strings"

	"github.com/caicloud/go-common/interfaces"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func JoinWebHooksPath(rootPath string, gvk schema.GroupVersionResource,
	opType arv1b1.OperationType) *string {
	result := path.Join(rootPath, gvk.Group, gvk.Version, gvk.Resource, strings.ToLower(string(opType)))
	return &result
}

func JoinObjectName(array ...string) string {
	vec := make([]string, 0, len(array))
	for _, s := range array {
		if s == "" {
			continue
		}
		vec = append(vec, s)
	}
	return strings.ToLower(strings.Join(vec, "."))
}

func RemoveObjectAnno(obj metav1.Object, key string) {
	if interfaces.IsNil(obj) {
		return
	}
	anno := obj.GetAnnotations()
	if len(anno) == 0 {
		return
	}
	delete(anno, key)
	obj.SetAnnotations(anno)
}
