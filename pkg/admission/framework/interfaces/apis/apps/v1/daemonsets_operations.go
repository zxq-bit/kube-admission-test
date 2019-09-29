package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (p *DaemonSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *DaemonSetProcessor) DoWithTracing(in *appsv1.DaemonSet) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(in)
	})
}

func (c *DaemonSetConfig) Register(opType arv1b1.OperationType, ps ...*DaemonSetProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]*DaemonSetProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]*DaemonSetProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		logPrefix := fmt.Sprintf("appsv1.DaemonSet[%v][%d][%s]", opType, i, p.Name)
		if e := p.Validate(); e != nil {
			log.Errorf("%s processor register failed, %v", logPrefix, e)
			continue
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
		log.Infof("%s processor register done", logPrefix)
	}
}

func (c *DaemonSetConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutSecondsMap == nil {
		c.TimeoutSecondsMap = make(map[arv1b1.OperationType]int32, 1)
	}
	c.TimeoutSecondsMap[opType] = int32(timeout / time.Second)
}

func (c *DaemonSetConfig) ToConfig(filter processor.MetadataFilter) (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: daemonsetsGVR,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetDaemonSetFromRawExtension(raw)
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
		ps = FilterDaemonSetProcessors(ps, filter)
		if len(ps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = CombineDaemonSetProcessors(ps)
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}

func GetDaemonSetFromRawExtension(raw *runtime.RawExtension) (*appsv1.DaemonSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != daemonsetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), daemonsetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.DaemonSet); obj != nil {
			return obj, nil
		}
	}
	parsed := &appsv1.DaemonSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

func FilterDaemonSetProcessors(in []*DaemonSetProcessor, filter processor.MetadataFilter) (out []*DaemonSetProcessor) {
	if filter == nil {
		return in
	}
	for _, p := range in {
		if p != nil && filter(&p.Metadata) {
			out = append(out, p)
		}
	}
	return out
}

func CombineDaemonSetProcessors(ps []*DaemonSetProcessor) util.Review {
	return func(ctx context.Context, in runtime.Object) (err error) {
		// check
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(*appsv1.DaemonSet)
		if obj == nil {
			return fmt.Errorf("not appsv1.DaemonSet")
		}
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", daemonsetsGVR.Group, daemonsetsGVR.Version, daemonsetsGVR.Resource)
		}
		// execute processors
		for i, p := range ps {
			// check ignore
			if ignoreReason := p.Metadata.GetObjectFilter()(obj); ignoreReason != nil {
				log.Infof("%s skip for %s", p.Name, *ignoreReason)
				continue
			}
			// do review
			var (
				logPrefix = logBase + fmt.Sprintf("[%d][%s]", i, p.Name)
				cost      time.Duration
			)
			select {
			case <-ctx.Done():
				err = errors.ErrContextEnded
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate:
					cost, err = p.DoWithTracing(obj.DeepCopy())
				case constants.ProcessorTypeMutate:
					cost, err = p.DoWithTracing(obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if err != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, err)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	}
}
