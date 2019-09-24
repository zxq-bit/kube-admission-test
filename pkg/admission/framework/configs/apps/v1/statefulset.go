package v1

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	statefulsetGRV = appsv1.SchemeGroupVersion.WithResource("statefulset")
	// gvk = appsv1.SchemeGroupVersion.WithKind(reflect.TypeOf(&appsv1.StatefulSet{}).Name())
)

type StatefulSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.StatefulSet) (err error)
}

type StatefulSetConfig struct {
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]StatefulSetProcessor
}

func (p *StatefulSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *StatefulSetConfig) Register(opType arv1b1.OperationType, ps ...StatefulSetProcessor) (errs []error) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]StatefulSetProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]StatefulSetProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if e := p.Validate(); e != nil {
			errs = append(errs, fmt.Errorf("[%d]%v", i, e))
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
	}
	return
}

func (c *StatefulSetConfig) ToConfig() (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: statefulsetGRV,
		ProcessorsMap:        make(map[arv1b1.OperationType][]processor.Processor, len(c.ProcessorsMap)),
	}
	for opType, ps := range c.ProcessorsMap {
		if len(ps) == 0 {
			continue
		}
		ops := make([]processor.Processor, 0, len(ps))
		for i := range ps {
			p := &ps[i]
			ops = append(ops, processor.Processor{
				Metadata: p.Metadata,
				Review: func(obj runtime.Object) (err error) {
					in := obj.(*appsv1.StatefulSet)
					if in == nil {
						err = fmt.Errorf("%s failed for input is nil", p.Name)
					} else {
						err = p.Review(in)
					}
					return err
				},
			})
		}
		if len(ops) > 0 {
			out.ProcessorsMap[opType] = ops
		}
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}
