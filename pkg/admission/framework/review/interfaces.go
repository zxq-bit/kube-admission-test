package review

import (
	"context"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RawExtensionParser func(raw *runtime.RawExtension) (runtime.Object, error)

type Reviewer interface {
	Register(processor interface{}) error
	DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error)
	IsEmpty() bool
}

type Handler interface {
	Key() string
	IsEmpty() bool
	SetTimeout(timeoutSecond int32) error
	Register(processor interface{}) error
	ToWebHook(svcNamespace, svcName string, svcCA []byte) arv1b1.Webhook
	ToNirvanaDescriptors() definition.Descriptor
}

func HandlerKey(gvr schema.GroupVersionResource, opType arv1b1.OperationType) string {
	return util.JoinObjectName(
		gvr.Group,
		gvr.Version,
		gvr.Resource,
		string(opType),
	)
}
