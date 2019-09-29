package util

import (
	"context"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"

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

func SetContextLogBase(ctx context.Context, logBase string) context.Context {
	return context.WithValue(ctx, constants.ContextKeyLogBase, logBase)
}

func GetContextLogBase(ctx context.Context) string {
	if interfaces.IsNil(ctx) {
		return ""
	}
	raw := ctx.Value(constants.ContextKeyLogBase)
	if interfaces.IsNil(raw) {
		return ""
	}
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	return s
}
