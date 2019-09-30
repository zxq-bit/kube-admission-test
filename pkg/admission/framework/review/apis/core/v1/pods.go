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
	review.RegisterHandler(podsGVR, NewPodReview)
}

type PodProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Review do review, return error if should stop
	Review func(ctx context.Context, in *corev1.Pod) (err error)
}

type PodReviewer struct {
	processors []*PodProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *PodProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *PodProcessor) DoWithTracing(ctx context.Context, in *corev1.Pod) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(ctx, in)
	})
}

// reviewer

func NewPodReview(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		podsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return podsRawExtensionParser(raw)
		},
		&PodReviewer{},
	)
}

func (r *PodReviewer) IsEmpty() bool {
	return len(r.processors) == 0
}

func (r *PodReviewer) Register(in interface{}) error {
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
	r.processors = append(r.processors, p)
	r.objFilters = append(r.objFilters, p.GetObjectFilter())
	return nil
}

func (r *PodReviewer) DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// check
		if interfaces.IsNil(in) {
			return errors.ErrNilRuntimeObject
		}
		obj := in.(*corev1.Pod)
		if obj == nil {
			return errors.ErrRuntimeObjectBadType
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", podsGVR.Group, podsGVR.Version, podsGVR.Resource)
			if opType := util.GetContextOpType(ctx); opType != "" {
				logBase += fmt.Sprintf("[%v]", opType)
			}
		}
		// execute processors
		for i, p := range r.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.Name)
			// check ignore
			if ignoreReason := r.objFilters[i](obj); ignoreReason != nil {
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

func podsRawExtensionParser(raw *runtime.RawExtension) (*corev1.Pod, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != podsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), podsGVK.String())
		}
		if obj := raw.Object.(*corev1.Pod); obj != nil {
			return obj, nil
		}
	}
	parsed := &corev1.Pod{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}