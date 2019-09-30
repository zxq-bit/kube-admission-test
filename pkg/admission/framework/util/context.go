package util

import (
	"context"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"

	"github.com/caicloud/go-common/interfaces"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
)

func getContextValue(ctx context.Context, k constants.ContextKey) string {
	if interfaces.IsNil(ctx) {
		return ""
	}
	raw := ctx.Value(k)
	if raw == nil {
		return ""
	}
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	return s
}

func SetContextLogBase(ctx context.Context, logBase string) context.Context {
	return context.WithValue(ctx, constants.ContextKeyLogBase, logBase)
}

func GetContextLogBase(ctx context.Context) string {
	return getContextValue(ctx, constants.ContextKeyLogBase)
}

func SetContextOpType(ctx context.Context, opType arv1b1.OperationType) context.Context {
	return context.WithValue(ctx, constants.ContextKeyOpType, string(opType))
}

func GetContextOpType(ctx context.Context) arv1b1.OperationType {
	return arv1b1.OperationType(
		getContextValue(ctx, constants.ContextKeyOpType),
	)
}
