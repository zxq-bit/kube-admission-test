package review

import (
	"context"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RawExtensionParser parse runtime.Object from runtime.RawExtension, implement by kinds
// inner interface for efficiency
type RawExtensionParser func(raw *runtime.RawExtension) (runtime.Object, error)

// Reviewer focus on review working
// inner interface to connect GVK processors with Handler
type Reviewer interface {
	Register(processor interface{}) error
	DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke *errors.StatusError)
	IsEmpty() bool
}

// Handler contains Reviewer, do around work for review
// eg: make http serve, webhook
type Handler interface {
	Key() string
	IsEmpty() bool
	SetTimeout(timeoutSecond int32) error
	Register(processor interface{}) error
	ToWebHook(svcNamespace, svcName string, svcCA []byte) arv1b1.Webhook
	ToNirvanaDescriptor() definition.Descriptor
}

// HandlerKey make key for handler by GroupVersionResource and OperationType
func HandlerKey(gvr schema.GroupVersionResource, opType arv1b1.OperationType) string {
	return util.JoinObjectName(
		gvr.Group,
		gvr.Version,
		gvr.Resource,
		string(opType),
	)
}
