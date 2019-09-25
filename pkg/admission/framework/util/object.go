package util

import (
	"strings"

	"github.com/caicloud/go-common/interfaces"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
