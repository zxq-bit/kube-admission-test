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
	deploymentsGVR = appsv1.SchemeGroupVersion.WithResource("deployments")
	deploymentsGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
)

func GetDeploymentGVR() schema.GroupVersionResource { return deploymentsGVR }
func GetDeploymentGVK() schema.GroupVersionKind     { return deploymentsGVK }

func init() {
	review.RegisterHandler(deploymentsGVR, NewDeploymentReview)
}

type DeploymentProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Review do review, return error if should stop
	Review func(ctx context.Context, in *appsv1.Deployment) (err error)
}

type DeploymentReviewer struct {
	processors []*DeploymentProcessor
	objFilters []util.ObjectIgnoreFilter
}

// processor

func (p *DeploymentProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (p *DeploymentProcessor) DoWithTracing(ctx context.Context, in *appsv1.Deployment) (cost time.Duration, err error) {
	return p.Tracer.DoWithTracing(func() error {
		return p.Review(ctx, in)
	})
}

// reviewer

func NewDeploymentReview(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		deploymentsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return deploymentsRawExtensionParser(raw)
		},
		&DeploymentReviewer{},
	)
}

func (r *DeploymentReviewer) IsEmpty() bool {
	return len(r.processors) == 0
}

func (r *DeploymentReviewer) Register(in interface{}) error {
	getProcessor := func(v interface{}) *DeploymentProcessor {
		if v == nil {
			return nil
		}
		return v.(*DeploymentProcessor)
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

func (r *DeploymentReviewer) DoReview(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, err error) {
	return tracer.DoWithTracing(func() (err error) {
		// check
		if interfaces.IsNil(in) {
			return errors.ErrNilRuntimeObject
		}
		obj := in.(*appsv1.Deployment)
		if obj == nil {
			return errors.ErrRuntimeObjectBadType
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		if logBase == "" {
			logBase = fmt.Sprintf("[%v/%v/%v]", deploymentsGVR.Group, deploymentsGVR.Version, deploymentsGVR.Resource)
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

func deploymentsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.Deployment, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != deploymentsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), deploymentsGVK.String())
		}
		if obj := raw.Object.(*appsv1.Deployment); obj != nil {
			return obj, nil
		}
	}
	parsed := &appsv1.Deployment{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}
