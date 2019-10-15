package util

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"

	"github.com/caicloud/go-common/interfaces"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
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

func GetContextOpType(ctx context.Context) arv1b1.OperationType {
	ar := getContextAdmissionRequest(ctx)
	if ar == nil {
		return ""
	}
	return arv1b1.OperationType(ar.Operation)
}

func GetContextOldObject(ctx context.Context) *runtime.RawExtension {
	ar := getContextAdmissionRequest(ctx)
	if ar == nil {
		return nil
	}
	return ar.OldObject.DeepCopy()
}

func SetContextAdmissionRequest(ctx context.Context, aReq *admissionv1b1.AdmissionRequest) context.Context {
	return context.WithValue(ctx, constants.ContextKeyAdmissionRequest, aReq)
}

func GetContextAdmissionRequest(ctx context.Context) *admissionv1b1.AdmissionRequest {
	if ar := getContextAdmissionRequest(ctx); ar != nil {
		return ar.DeepCopy()
	}
	return nil
}

func getContextAdmissionRequest(ctx context.Context) *admissionv1b1.AdmissionRequest {
	if interfaces.IsNil(ctx) {
		return nil
	}
	raw := ctx.Value(constants.ContextKeyAdmissionRequest)
	if raw == nil {
		return nil
	}
	return raw.(*admissionv1b1.AdmissionRequest)
}
