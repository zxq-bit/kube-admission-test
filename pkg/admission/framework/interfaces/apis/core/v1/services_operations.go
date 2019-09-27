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

func (p *ServiceProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *ServiceConfig) Register(opType arv1b1.OperationType, ps ...*ServiceProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]ServiceProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]ServiceProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		logPrefix := fmt.Sprintf("corev1.Service[%v][%d][%s]", opType, i, p.Name)
		if e := p.Validate(); e != nil {
			log.Errorf("%s processor register failed, %v", logPrefix, e)
			continue
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], *p)
		log.Infof("%s processor register done", logPrefix)
	}
}

func (c *ServiceConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutSecondsMap == nil {
		c.TimeoutSecondsMap = make(map[arv1b1.OperationType]int32, 1)
	}
	c.TimeoutSecondsMap[opType] = int32(timeout / time.Second)
}

func (c *ServiceConfig) ToConfig(filter processor.MetadataFilter) (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: servicesGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetServiceFromRawExtension(raw)
			if e != nil {
				return nil, e
			}
			return obj, nil
		},
		TimeoutSecondsMap: c.TimeoutSecondsMap,
		ProcessorsMap:     make(map[arv1b1.OperationType]util.Review, len(c.ProcessorsMap)),
	}
	if out.TimeoutSecondsMap == nil {
		out.TimeoutSecondsMap = map[arv1b1.OperationType]int32{}
	}
	for opType, ps := range c.ProcessorsMap {
		ps = FilterServiceProcessors(ps, filter)
		if len(ps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = CombineServiceProcessors(ps)
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}

func GetServiceFromRawExtension(raw *runtime.RawExtension) (*corev1.Service, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != servicesGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), servicesGVK.String())
		}
		if obj := raw.Object.(*corev1.Service); obj != nil {
			return obj, nil
		}
	}
	parsed := &corev1.Service{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

func FilterServiceProcessors(in []ServiceProcessor, filter processor.MetadataFilter) (out []ServiceProcessor) {
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

func CombineServiceProcessors(ps []ServiceProcessor) util.Review {
	return func(ctx context.Context, in runtime.Object) (err error) {
		// check
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(*corev1.Service)
		if obj == nil {
			return fmt.Errorf("not corev1.Service")
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
