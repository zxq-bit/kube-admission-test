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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	daemonsetsGVR = appsv1.SchemeGroupVersion.WithResource("daemonsets")
	daemonsetsGVK = appsv1.SchemeGroupVersion.WithKind("DaemonSet")
)

func GetDaemonSetGVR() schema.GroupVersionResource { return daemonsetsGVR }
func GetDaemonSetGVK() schema.GroupVersionKind     { return daemonsetsGVK }

func init() {
	review.RegisterHandlerMaker(daemonsetsGVR, NewDaemonSetHandler)
}

type DaemonSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.DaemonSet) errors.APIStatus
}

type DaemonSetHandler struct {
	processors []*DaemonSetProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *DaemonSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *DaemonSetProcessor) DoWithTracing(ctx context.Context, in *appsv1.DaemonSet) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// handler

func NewDaemonSetHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		daemonsetsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return daemonsetsRawExtensionParser(raw)
		},
		&DaemonSetHandler{},
	)
}

func (h *DaemonSetHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *DaemonSetHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *DaemonSetProcessor {
		if v == nil {
			return nil
		}
		return v.(*DaemonSetProcessor)
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

func (h *DaemonSetHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.DaemonSet {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.DaemonSet)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldDaemonSet(ctx)
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
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *appsv1.DaemonSet
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func daemonsetsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.DaemonSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != daemonsetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), daemonsetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.DaemonSet); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.DaemonSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldDaemonSet get DaemonSet old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldDaemonSet(ctx context.Context) (*appsv1.DaemonSet, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return daemonsetsRawExtensionParser(raw)
}
