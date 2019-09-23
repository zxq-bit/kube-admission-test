package v1

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	configmapGRV = corev1.SchemeGroupVersion.WithResource("configmap")
	// gvk = corev1.SchemeGroupVersion.WithKind(reflect.TypeOf(&corev1.ConfigMap{}).Name())
)

type ConfigMapProcessor struct {
	Name string
	// IgnoreSetting set namespaces and annotations that will ignore this processor
	IgnoreSetting *util.IgnoreSetting
	// Type Validate or Mutate, decide weather to allow input object changes
	Type util.ProcessorType
	// Review do review, return error if should stop
	Review func(in *corev1.ConfigMap) (err error)
}

type ConfigMapConfig struct {
	// ModelName describe the model of this config, like app or resource
	ModelName string
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]ConfigMapProcessor
}

func (p *ConfigMapProcessor) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("empty processor name")
	}
	if p.Type != util.ProcessorTypeValidate && p.Type != util.ProcessorTypeMutate {
		return fmt.Errorf("%v invalid processor type %v", p.Name, p.Type)
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *ConfigMapConfig) Register(opType arv1b1.OperationType, ps ...ConfigMapProcessor) (errs []error) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]ConfigMapProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]ConfigMapProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if e := p.Validate(); e != nil {
			errs = append(errs, fmt.Errorf("[%d]%v", i, e))
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
	}
	return
}

func (c *ConfigMapConfig) ToConfig() (out *util.Config) {
	out = &util.Config{
		ModelName:            c.ModelName,
		GroupVersionResource: configmapGRV,
		ProcessorsMap:        make(map[arv1b1.OperationType][]util.Processor, len(c.ProcessorsMap)),
	}
	for opType, pps := range c.ProcessorsMap {
		if len(pps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = make([]util.Processor, 0, len(pps))
		for i := range pps {
			p := &pps[i]
			out.ProcessorsMap[opType] = append(out.ProcessorsMap[opType], util.Processor{
				Name:          p.Name,
				IgnoreSetting: p.IgnoreSetting,
				Type:          p.Type,
				Review: func(obj runtime.Object) (err error) {
					in := obj.(*corev1.ConfigMap)
					if in == nil {
						err = fmt.Errorf("%s failed for input is nil", p.Name)
					} else {
						err = p.Review(in)
					}
					return err
				},
			})
		}
	}
	return out
}
