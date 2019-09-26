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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (p *StatefulSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *StatefulSetConfig) Register(opType arv1b1.OperationType, ps ...*StatefulSetProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]StatefulSetProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]StatefulSetProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		if e := p.Validate(); e != nil {
			log.Errorf("appsv1.StatefulSet processor register failed for [%d.%s], %v", i, p.Name, e)
			continue
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], *p)
		log.Infof("appsv1.StatefulSet processor register done for [%d.%s], %v", i, p.Name)
	}
	return
}

func (c *StatefulSetConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutMap == nil {
		c.TimeoutMap = make(map[arv1b1.OperationType]time.Duration, 1)
	}
	c.TimeoutMap[opType] = timeout
}

func (c *StatefulSetConfig) ToConfig(filter processor.MetadataFilter) (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: statefulsetGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetStatefulSetFromRawExtension(raw)
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
		ps = FilterStatefulSetProcessors(ps, filter)
		if len(ps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = CombineStatefulSetProcessors(ps)
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}

func GetStatefulSetFromRawExtension(raw *runtime.RawExtension) (*appsv1.StatefulSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != statefulsetGVK {
		return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), statefulsetGVK.String())
	}
	if obj := raw.Object.(*appsv1.StatefulSet); obj != nil {
		return obj, nil
	}
	parsed := &appsv1.StatefulSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

func FilterStatefulSetProcessors(in []StatefulSetProcessor, filter processor.MetadataFilter) (out []StatefulSetProcessor) {
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

func CombineStatefulSetProcessors(ps []StatefulSetProcessor) util.Review {
	return func(ctx context.Context, in runtime.Object) (err error) {
		// check
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(*appsv1.StatefulSet)
		if obj == nil {
			return fmt.Errorf("not appsv1.StatefulSet")
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
