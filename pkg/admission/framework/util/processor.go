package util

import (
	"context"
	"fmt"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ReviewFunc func(in runtime.Object) (err error)
type ReviewFuncWithContext func(ctx context.Context, in runtime.Object) (err error)

type Processor struct {
	Name string
	// IgnoreSetting set namespaces and annotations that will ignore this processor
	IgnoreSetting *IgnoreSetting
	// Type Validate or Mutate, decide weather to allow input object changes
	Type ProcessorType
	// Review do review, return error if should stop
	Review ReviewFunc
}

func (p *Processor) FilterObject(obj metav1.Object) *string {
	if p.IgnoreSetting == nil {
		return nil
	}
	return p.IgnoreSetting.GetObjectFilter(AnnoKeyAdmissionIgnore)(obj)
}

type Config struct {
	// ModelName describe the model of this config, like app or resource
	ModelName string
	// Group+Version+Resource
	GroupVersionResource schema.GroupVersionResource
	// ProcessorsMap map processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]Processor
}

func (c *Config) ToMutatingWebHook(svcConfig *ServiceConfig) (re *arv1b1.MutatingWebhookConfiguration) {
	gvr := c.GroupVersionResource
	re = &arv1b1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: makeMutatingWebHookName(gvr.Group, gvr.Version, gvr.Resource),
		},
		Webhooks: make([]arv1b1.Webhook, 0, len(c.ProcessorsMap)),
	}
	for opType, processor := range c.ProcessorsMap {
		if processor != nil {
			re.Webhooks = append(re.Webhooks, MakeMutatingWebHook(
				svcConfig,
				gvr,
				opType,
			))
		}
	}
	if len(re.Webhooks) == 0 {
		return nil
	}
	return re
}

func (c *Config) ToNirvanaDescriptors(svcConfig *ServiceConfig) (re []definition.Descriptor) {
	if len(c.ProcessorsMap) == 0 {
		return
	}
	gvr := c.GroupVersionResource
	for _, opType := range OperationTypes {
		processors := c.ProcessorsMap[opType]
		if len(processors) == 0 {
			continue
		}
		re = append(re, MakeNirvanaDescriptor(
			svcConfig,
			gvr,
			opType,
			CombineProcessors(processors),
		))
	}
	return
}

func CombineProcessors(processors []Processor) ReviewFuncWithContext {
	ps := filterOutInvalidProcessors(processors)
	return func(ctx context.Context, in runtime.Object) (err error) {
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(metav1.Object)
		if interfaces.IsNil(obj) {
			return fmt.Errorf("not metav1 object")
		}
		// always return a not nil `out`, if out is nil, use in
		for _, p := range ps {
			// check ignore
			if ignoreReason := p.FilterObject(obj); ignoreReason != nil {
				log.Infof("%s skip for %s", p.Name, *ignoreReason)
				continue
			}
			// do review
			select {
			case <-ctx.Done():
				err = fmt.Errorf("processor chain not finished correctly, context ended")
			default:
				switch p.Type {
				case ProcessorTypeValidate:
					err = p.Review(in.DeepCopyObject())
				case ProcessorTypeMutate:
					err = p.Review(in)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if err != nil {
				break
			}
		}
		return
	}
}

func filterOutInvalidProcessors(ins []Processor) (outs []Processor) {
	for _, p := range ins {
		if p.Review == nil {
			continue
		}
		outs = append(outs, p)
	}
	return
}
