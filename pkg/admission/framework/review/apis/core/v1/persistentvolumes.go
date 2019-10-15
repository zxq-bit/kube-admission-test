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
	persistentvolumesGVR = corev1.SchemeGroupVersion.WithResource("persistentvolumes")
	persistentvolumesGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolume")
)

func GetPersistentVolumeGVR() schema.GroupVersionResource { return persistentvolumesGVR }
func GetPersistentVolumeGVK() schema.GroupVersionKind     { return persistentvolumesGVK }

func init() {
	review.RegisterHandlerMaker(persistentvolumesGVR, NewPersistentVolumeHandler)
}

type PersistentVolumeProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Review do review, return error if should stop
	Review func(ctx context.Context, in *corev1.PersistentVolume) (err error)
}

type PersistentVolumeHandler struct {
	processors []*PersistentVolumeProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *PersistentVolumeProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Key())
	}
	return nil
}

func (p *PersistentVolumeProcessor) DoWithTracing(ctx context.Context, in *corev1.PersistentVolume) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(ctx, in)
	})
}

// reviewer

func NewPersistentVolumeHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		persistentvolumesGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return persistentvolumesRawExtensionParser(raw)
		},
		&PersistentVolumeHandler{},
	)
}

func (h *PersistentVolumeHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PersistentVolumeHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PersistentVolumeProcessor {
		if v == nil {
			return nil
		}
		return v.(*PersistentVolumeProcessor)
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

func (h *PersistentVolumeHandler) DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.PersistentVolume {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.PersistentVolume)
		}()
		toFilter := obj
		if toFilter == nil {
			toFilter, err = GetContextOldPersistentVolume(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoReview failed, %v", logBase, err)
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
					var toValidate *corev1.PersistentVolume
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

func persistentvolumesRawExtensionParser(raw *runtime.RawExtension) (*corev1.PersistentVolume, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != persistentvolumesGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), persistentvolumesGVK.String())
		}
		if obj := raw.Object.(*corev1.PersistentVolume); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.PersistentVolume{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldPersistentVolume get PersistentVolume old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldPersistentVolume(ctx context.Context) (*corev1.PersistentVolume, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return persistentvolumesRawExtensionParser(raw)
}
