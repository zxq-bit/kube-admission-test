package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (p *PodProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *PodConfig) Register(opType arv1b1.OperationType, ps ...*PodProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]PodProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]PodProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		if e := p.Validate(); e != nil {
			log.Errorf("corev1.Pod processor register failed for [%d.%s], %v", i, p.Name, e)
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], *p)
	}
	return
}

func (c *PodConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutMap == nil {
		c.TimeoutMap = make(map[arv1b1.OperationType]time.Duration, 1)
	}
	c.TimeoutMap[opType] = timeout
}

func (c *PodConfig) ToConfig(filter processor.MetadataFilter) (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: podsGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetPodFromRawExtension(raw)
			if e != nil {
				return nil, e
			}
			return obj, nil
		},
		TimeoutMap:    c.TimeoutMap,
		ProcessorsMap: make(map[arv1b1.OperationType]util.Review, len(c.ProcessorsMap)),
	}
	if out.TimeoutMap == nil {
		out.TimeoutMap = map[arv1b1.OperationType]time.Duration{}
	}
	for opType, ps := range c.ProcessorsMap {
		ps = FilterPodProcessors(ps, filter)
		if len(ps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = CombinePodProcessors(ps)
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

func FilterPodProcessors(in []PodProcessor, filter processor.MetadataFilter) (out []PodProcessor) {
	if filter == nil {
		return in
	}
	for _, p := range in {
		if filter(&p.Metadata) {
			out = append(out, p)
		}
	}
	return out
}

func CombinePodProcessors(ps []PodProcessor) util.Review {
	return func(ctx context.Context, in runtime.Object) (err error) {
		// check
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(*corev1.Pod)
		if obj == nil {
			return fmt.Errorf("not corev1.Pod")
		}
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for _, p := range ps {
			// check ignore
			if ignoreReason := p.Metadata.GetObjectFilter()(obj); ignoreReason != nil {
				log.Infof("%s skip for %s", p.Name, *ignoreReason)
				continue
			}
			// do review
			select {
			case <-ctx.Done():
				err = fmt.Errorf("processor chain not finished correctly, context ended")
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate:
					err = p.Review(obj.DeepCopy())
				case constants.ProcessorTypeMutate:
					err = p.Review(obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if err != nil {
				log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				break
			}
		}
		return
	}
}
