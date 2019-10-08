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
	persistentvolumeclaimsGVR = corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims")
	persistentvolumeclaimsGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")
)

func GetPersistentVolumeClaimGVR() schema.GroupVersionResource { return persistentvolumeclaimsGVR }
func GetPersistentVolumeClaimGVK() schema.GroupVersionKind     { return persistentvolumeclaimsGVK }

func init() {
	review.RegisterHandlerMaker(persistentvolumeclaimsGVR, NewPersistentVolumeClaimHandler)
}

type PersistentVolumeClaimProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Review do review, return error if should stop
	Review func(ctx context.Context, in *corev1.PersistentVolumeClaim) (err error)
}

type PersistentVolumeClaimHandler struct {
	processors []*PersistentVolumeClaimProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *PersistentVolumeClaimProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *PersistentVolumeClaimProcessor) DoWithTracing(ctx context.Context, in *corev1.PersistentVolumeClaim) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(ctx, in)
	})
}

// reviewer

func NewPersistentVolumeClaimHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		persistentvolumeclaimsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return persistentvolumeclaimsRawExtensionParser(raw)
		},
		&PersistentVolumeClaimHandler{},
	)
}

func (h *PersistentVolumeClaimHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PersistentVolumeClaimHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PersistentVolumeClaimProcessor {
		if v == nil {
			return nil
		}
		return v.(*PersistentVolumeClaimProcessor)
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

func (h *PersistentVolumeClaimHandler) DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// check
		if interfaces.IsNil(in) {
			return errors.ErrNilRuntimeObject
		}
		obj := in.(*corev1.PersistentVolumeClaim)
		if obj == nil {
			return errors.ErrRuntimeObjectBadType
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", persistentvolumeclaimsGVR.Group, persistentvolumeclaimsGVR.Version, persistentvolumeclaimsGVR.Resource)
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

func persistentvolumeclaimsRawExtensionParser(raw *runtime.RawExtension) (*corev1.PersistentVolumeClaim, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != persistentvolumeclaimsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), persistentvolumeclaimsGVK.String())
		}
		if obj := raw.Object.(*corev1.PersistentVolumeClaim); obj != nil {
			return obj, nil
		}
	}
	parsed := &corev1.PersistentVolumeClaim{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}
