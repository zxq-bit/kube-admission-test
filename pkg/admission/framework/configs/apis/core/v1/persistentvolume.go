package v1

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	persistentvolumeGRV = corev1.SchemeGroupVersion.WithResource("persistentvolume")
	persistentvolumeGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolume")
)

type PersistentVolumeProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.PersistentVolume) (err error)
}

type PersistentVolumeConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PersistentVolumeProcessor
}

func (p *PersistentVolumeProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *PersistentVolumeConfig) Register(opType arv1b1.OperationType, ps ...*PersistentVolumeProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]PersistentVolumeProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]PersistentVolumeProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		if e := p.Validate(); e != nil {
			log.Errorf("corev1.PersistentVolume processor register failed for [%d.%s], %v", i, p.Name, e)
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], *p)
	}
	return
}

func (c *PersistentVolumeConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutMap == nil {
		c.TimeoutMap = make(map[arv1b1.OperationType]time.Duration, 1)
	}
	c.TimeoutMap[opType] = timeout
}

func (c *PersistentVolumeConfig) ToConfig() (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: persistentvolumeGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetPersistentVolumeFromRawExtension(raw)
			if e != nil {
				return nil, e
			}
			return obj, nil
		},
		TimeoutMap:    c.TimeoutMap,
		ProcessorsMap: make(map[arv1b1.OperationType][]processor.Processor, len(c.ProcessorsMap)),
	}
	if out.TimeoutMap == nil {
		out.TimeoutMap = map[arv1b1.OperationType]time.Duration{}
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
					in := obj.(*corev1.PersistentVolume)
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

func GetPersistentVolumeFromRawExtension(raw *runtime.RawExtension) (*corev1.PersistentVolume, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != persistentvolumeGVK {
		return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), persistentvolumeGVK.String())
	}
	if obj := raw.Object.(*corev1.PersistentVolume); obj != nil {
		return obj, nil
	}
	parsed := &corev1.PersistentVolume{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}
