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
	replicasetsGVR = appsv1.SchemeGroupVersion.WithResource("replicasets")
	replicasetsGVK = appsv1.SchemeGroupVersion.WithKind("ReplicaSet")
)

func GetReplicaSetGVR() schema.GroupVersionResource { return replicasetsGVR }
func GetReplicaSetGVK() schema.GroupVersionKind     { return replicasetsGVK }

func init() {
	review.RegisterHandlerMaker(replicasetsGVR, NewReplicaSetHandler)
}

type ReplicaSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.ReplicaSet) (err error)
}

type ReplicaSetHandler struct {
	processors []*ReplicaSetProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *ReplicaSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *ReplicaSetProcessor) DoWithTracing(ctx context.Context, in *appsv1.ReplicaSet) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Admit(ctx, in)
	})
}

// handler

func NewReplicaSetHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		replicasetsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return replicasetsRawExtensionParser(raw)
		},
		&ReplicaSetHandler{},
	)
}

func (h *ReplicaSetHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *ReplicaSetHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *ReplicaSetProcessor {
		if v == nil {
			return nil
		}
		return v.(*ReplicaSetProcessor)
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

func (h *ReplicaSetHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.ReplicaSet {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.ReplicaSet)
		}()
		toFilter := obj
		if toFilter == nil {
			toFilter, err = GetContextOldReplicaSet(ctx)
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
					var toValidate *appsv1.ReplicaSet
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

func replicasetsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.ReplicaSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != replicasetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), replicasetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.ReplicaSet); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.ReplicaSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldReplicaSet get ReplicaSet old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldReplicaSet(ctx context.Context) (*appsv1.ReplicaSet, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return replicasetsRawExtensionParser(raw)
}
