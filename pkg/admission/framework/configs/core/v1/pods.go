package v1

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	podsGRV = corev1.SchemeGroupVersion.WithResource("pods")
	podsGVK = corev1.SchemeGroupVersion.WithKind(reflect.TypeOf(&corev1.Pod{}).Name())
)

type PodProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.Pod) (err error)
}

type PodConfig struct {
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PodProcessor
}

func (p *PodProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *PodConfig) Register(opType arv1b1.OperationType, ps ...PodProcessor) (errs []error) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]PodProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]PodProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if e := p.Validate(); e != nil {
			errs = append(errs, fmt.Errorf("[%d]%v", i, e))
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
	}
	return
}

func (c *PodConfig) ToConfig() (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: podsGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetPodFromRawExtension(raw)
			if e != nil {
				return nil, e
			}
			return obj, nil
		},
		ProcessorsMap: make(map[arv1b1.OperationType][]processor.Processor, len(c.ProcessorsMap)),
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
					in := obj.(*corev1.Pod)
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

func GetPodFromRawExtension(raw *runtime.RawExtension) (*corev1.Pod, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != podsGVK {
		return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), podsGVK.String())
	}
	if obj := raw.Object.(*corev1.Pod); obj != nil {
		return obj, nil
	}
	parsed := &corev1.Pod{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}
