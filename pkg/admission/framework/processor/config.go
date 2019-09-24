package processor

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/options"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RawExtensionParser func(raw *runtime.RawExtension) (runtime.Object, error)

type Config struct {
	// Group+Version+Resource
	GroupVersionResource schema.GroupVersionResource
	// RawExtensionParser RawExtensionParser
	RawExtensionParser RawExtensionParser
	// ProcessorsMap map processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]Processor
}

func (c *Config) ToMutatingWebHook(opt *options.StartOptions) (re *arv1b1.MutatingWebhookConfiguration) {
	gvr := c.GroupVersionResource
	re = &arv1b1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.JoinObjectName(gvr.Group, gvr.Version, gvr.Resource),
		},
		Webhooks: make([]arv1b1.Webhook, 0, len(c.ProcessorsMap)),
	}
	for _, opType := range constants.OperationTypes {
		ps := opt.FilterProcessorsByModel(c.ProcessorsMap[opType])
		if len(ps) == 0 {
			continue
		}
		re.Webhooks = append(re.Webhooks, makeMutatingWebHook(
			opt,
			gvr,
			opType,
		))
	}
	if len(re.Webhooks) == 0 {
		return nil
	}
	return re
}

func (c *Config) ToNirvanaDescriptors(opt *options.StartOptions) (re []definition.Descriptor) {
	if len(c.ProcessorsMap) == 0 {
		return
	}
	gvr := c.GroupVersionResource
	for _, opType := range constants.OperationTypes {
		ps := opt.FilterProcessorsByModel(c.ProcessorsMap[opType])
		if len(ps) == 0 {
			continue
		}
		re = append(re, makeNirvanaDescriptor(
			opt.APIRootPath,
			gvr,
			opType,
			CombineProcessors(ps),
		))
	}
	return
}
