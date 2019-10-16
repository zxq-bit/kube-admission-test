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
	podsGVR = corev1.SchemeGroupVersion.WithResource("pods")
	podsGVK = corev1.SchemeGroupVersion.WithKind("Pod")
)

func GetPodGVR() schema.GroupVersionResource { return podsGVR }
func GetPodGVK() schema.GroupVersionKind     { return podsGVK }

func init() {
	review.RegisterHandlerMaker(podsGVR, NewPodHandler)
}

type PodProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.Pod) (err error)
}

type PodHandler struct {
	processors []*PodProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *PodProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *PodProcessor) DoWithTracing(ctx context.Context, in *corev1.Pod) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Admit(ctx, in)
	})
}

// handler

func NewPodHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		podsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return podsRawExtensionParser(raw)
		},
		&PodHandler{},
	)
}

func (h *PodHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PodHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PodProcessor {
		if v == nil {
			return nil
		}
		return v.(*PodProcessor)
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

func (h *PodHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.Pod {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.Pod)
		}()
		toFilter := obj
		if toFilter == nil {
			toFilter, err = GetContextOldPod(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
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
					var toValidate *corev1.Pod
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, err = p.DoWithTracing(ctx, toValidate)
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

func podsRawExtensionParser(raw *runtime.RawExtension) (*corev1.Pod, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != podsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), podsGVK.String())
		}
		if obj := raw.Object.(*corev1.Pod); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.Pod{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldPod get Pod old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldPod(ctx context.Context) (*corev1.Pod, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return podsRawExtensionParser(raw)
}
