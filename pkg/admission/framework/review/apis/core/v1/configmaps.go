package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/handler"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	configmapsGVR = corev1.SchemeGroupVersion.WithResource("configmaps")
	configmapsGVK = corev1.SchemeGroupVersion.WithKind("ConfigMap")
)

func GetConfigMapGVR() schema.GroupVersionResource { return configmapsGVR }
func GetConfigMapGVK() schema.GroupVersionKind     { return configmapsGVK }

func init() {
	review.RegisterHandlerMaker(configmapsGVR, NewConfigMapHandler)
}

type ConfigMapProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Review do review, return error if should stop
	Review func(ctx context.Context, in *corev1.ConfigMap) (err error)
}

type ConfigMapHandler struct {
	processors []*ConfigMapProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *ConfigMapProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *ConfigMapProcessor) DoWithTracing(ctx context.Context, in *corev1.ConfigMap) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(ctx, in)
	})
}

// reviewer

func NewConfigMapHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		configmapsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return configmapsRawExtensionParser(raw)
		},
		&ConfigMapHandler{},
	)
}

func (h *ConfigMapHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *ConfigMapHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *ConfigMapProcessor {
		if v == nil {
			return nil
		}
		return v.(*ConfigMapProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *ConfigMapHandler) DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// check
		if interfaces.IsNil(in) {
			return errors.ErrNilRuntimeObject
		}
		obj := in.(*corev1.ConfigMap)
		if obj == nil {
			return errors.ErrRuntimeObjectBadType
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", configmapsGVR.Group, configmapsGVR.Version, configmapsGVR.Resource)
			if opType := util.GetContextOpType(ctx); opType != "" {
				logBase += fmt.Sprintf("[%v]", opType)
			}
		}
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.Name)
			// check ignore
			if ignoreReason := h.objFilters[i](obj); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				err = errors.ErrContextEnded
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					cost, err = p.DoWithTracing(ctx, obj.DeepCopy())
				case constants.ProcessorTypeMutate:
					cost, err = p.DoWithTracing(ctx, obj)
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
	})
}

func configmapsRawExtensionParser(raw *runtime.RawExtension) (*corev1.ConfigMap, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != configmapsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), configmapsGVK.String())
		}
		if obj := raw.Object.(*corev1.ConfigMap); obj != nil {
			return obj, nil
		}
	}
	parsed := &corev1.ConfigMap{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldConfigMap get ConfigMap old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return configmapsRawExtensionParser(raw)
}
