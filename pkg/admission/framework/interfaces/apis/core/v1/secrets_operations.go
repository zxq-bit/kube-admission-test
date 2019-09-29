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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (p *SecretProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *SecretProcessor) DoWithTracing(in *corev1.Secret) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(in)
	})
}

func (c *SecretConfig) Register(opType arv1b1.OperationType, ps ...*SecretProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]*SecretProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]*SecretProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		logPrefix := fmt.Sprintf("corev1.Secret[%v][%d][%s]", opType, i, p.Name)
		if e := p.Validate(); e != nil {
			log.Errorf("%s processor register failed, %v", logPrefix, e)
			continue
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
		log.Infof("%s processor register done", logPrefix)
	}
}

func (c *SecretConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutSecondsMap == nil {
		c.TimeoutSecondsMap = make(map[arv1b1.OperationType]int32, 1)
	}
	c.TimeoutSecondsMap[opType] = int32(timeout / time.Second)
}

func (c *SecretConfig) ToConfig(filter processor.MetadataFilter) (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: secretsGVR,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetSecretFromRawExtension(raw)
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
		ps = FilterSecretProcessors(ps, filter)
		if len(ps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = CombineSecretProcessors(ps)
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}

func GetSecretFromRawExtension(raw *runtime.RawExtension) (*corev1.Secret, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != secretsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), secretsGVK.String())
		}
		if obj := raw.Object.(*corev1.Secret); obj != nil {
			return obj, nil
		}
	}
	parsed := &corev1.Secret{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

func FilterSecretProcessors(in []*SecretProcessor, filter processor.MetadataFilter) (out []*SecretProcessor) {
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

func CombineSecretProcessors(ps []*SecretProcessor) util.Review {
	return func(ctx context.Context, in runtime.Object) (err error) {
		// check
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(*corev1.Secret)
		if obj == nil {
			return fmt.Errorf("not corev1.Secret")
		}
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", secretsGVR.Group, secretsGVR.Version, secretsGVR.Resource)
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
