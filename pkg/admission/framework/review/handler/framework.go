package handler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/logger"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/go-common/nirvana/middleware"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// failurePolicyTypeIgnore = arv1b1.Ignore
	failurePolicyTypeFail = arv1b1.Fail
)

type framework struct {
	// Group+Version+Resource
	gvr schema.GroupVersionResource
	// OperationType
	opType arv1b1.OperationType
	// parser parser of RawExtension
	parser review.RawExtensionParser
	// timeoutSecond set total execute time of processors
	timeoutSecond int32
	// tracer do trace and save data
	tracer tracer.Tracer
	// reviewer do review
	reviewer review.Reviewer
}

func NewFramework(gvr schema.GroupVersionResource, opType arv1b1.OperationType,
	parser review.RawExtensionParser, reviewer review.Reviewer) (review.Handler, error) {
	if gvr.Empty() {
		return nil, errors.ErrEmptyGVR
	}
	if !util.IsOperationTypeLeague(opType) {
		return nil, errors.ErrBadOperationType
	}
	if parser == nil {
		return nil, errors.ErrNilRawExtensionParser
	}
	if interfaces.IsNil(reviewer) {
		return nil, errors.ErrNilReviewer
	}
	return &framework{
		gvr:      gvr,
		opType:   opType,
		parser:   parser,
		reviewer: reviewer,
	}, nil
}

func (r *framework) Key() string {
	return review.HandlerKey(r.gvr, r.opType)
}

func (r *framework) IsEmpty() bool {
	return r.reviewer.IsEmpty()
}

func (r *framework) SetTimeout(timeoutSecond int32) error {
	if timeoutSecond < 0 {
		return errors.ErrBadTimeoutSecond
	}
	r.timeoutSecond = timeoutSecond
	return nil
}

func (r *framework) Register(processor interface{}) error {
	return r.reviewer.Register(processor)
}

func (r *framework) ToWebHook(svcNamespace, svcName string, svcCA []byte) arv1b1.Webhook {
	var timeoutSeconds *int32
	if r.timeoutSecond > 0 {
		timeoutSeconds = &r.timeoutSecond
	}
	return arv1b1.Webhook{
		Name: r.Key(),
		Rules: []arv1b1.RuleWithOperations{
			{
				Operations: []arv1b1.OperationType{r.opType},
				Rule: arv1b1.Rule{
					APIGroups:   []string{r.gvr.Group},
					APIVersions: []string{r.gvr.Version},
					Resources:   []string{r.gvr.Resource},
				},
			},
		},
		ClientConfig: arv1b1.WebhookClientConfig{
			Service: &arv1b1.ServiceReference{
				Name:      svcName,
				Namespace: svcNamespace,
				Path:      util.JoinWebHooksPath(constants.APIRootPath, r.gvr, r.opType),
			},
			CABundle: svcCA,
		},
		TimeoutSeconds: timeoutSeconds,
		FailurePolicy:  &failurePolicyTypeFail,
	}
}

func (r *framework) ToNirvanaDescriptors() definition.Descriptor {
	// timeout
	middlewares := []definition.Middleware{
		logger.New(log.DefaultLogger()),
		tracer.New(&r.tracer),
	}
	if r.timeoutSecond > 0 {
		middlewares = append(middlewares, middleware.Timeout(time.Duration(r.timeoutSecond)*time.Second))
	}
	logBase := fmt.Sprintf("[%v/%v/%v][%v]", r.gvr.Group, r.gvr.Version, r.gvr.Resource, r.opType)
	// descriptor
	return definition.Descriptor{
		Path:        *util.JoinWebHooksPath(constants.APIRootPath, r.gvr, r.opType),
		Middlewares: middlewares,
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Description: fmt.Sprintf("do admission for %s", logBase),
				Consumes:    []string{definition.MIMEAll},
				Produces:    []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source:      definition.Body,
						Name:        "admissionReview",
						Description: "admission body",
						Operators: []definition.Operator{
							definition.OperatorFunc("AdmissionReview parse",
								func(ctx context.Context, name string, body []byte) (*admissionv1b1.AdmissionReview, error) {
									ar := admissionv1b1.AdmissionReview{}
									gvk := admissionv1b1.SchemeGroupVersion.WithKind(reflect.TypeOf(ar).Name())
									deserializer := scheme.Codecs.UniversalDeserializer()
									if _, _, err := deserializer.Decode(body, &gvk, &ar); err != nil {
										log.Errorf("%s decode AdmissionReview failed, %v", logBase, err)
										ar.Response = util.ToAdmissionFailedResponse("", err)
									} else {
										log.DefaultLogger().Infof("%s decode AdmissionReview done: %s", logBase, string(body))
									}
									return &ar, nil
								}),
						},
					},
				},
				Results: []definition.Result{
					definition.DataResultFor("admission response"),
				},
				Function: func(ctx context.Context, ar *admissionv1b1.AdmissionReview) *admissionv1b1.AdmissionReview {
					// check
					if ar.Response != nil {
						return ar
					}
					if ar.Request == nil {
						e := fmt.Errorf("empty AdmissionReview Request")
						log.Errorf("%s check AdmissionReview failed, %v", logBase, e)
						ar.Response = util.ToAdmissionFailedResponse("", e)
						return ar
					}
					// parse raw
					logPrefix := fmt.Sprintf("%s[%s/%s]", logBase, ar.Request.Namespace, ar.Request.Name)
					log.DefaultLogger().Infof("%s start", logPrefix)
					org, e := r.parser(&ar.Request.Object)
					if e != nil {
						log.Errorf("%s RawExtensionParser failed, %v", logPrefix, e)
						ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
						return ar
					}
					log.DefaultLogger().Infof("%s RawExtensionParser done", logPrefix)
					// context
					ctx = util.SetContextLogBase(ctx, logPrefix)
					ctx = util.SetContextOpType(ctx, r.opType)
					// do review
					obj := org.DeepCopyObject()
					cost, e := r.reviewer.DoReview(ctx, &r.tracer, obj)
					if e != nil {
						ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
						log.Errorf("%s[cost:%v] do review failed, %v", logPrefix, cost, e)
						return ar
					}
					ar.Response = util.ToAdmissionPassResponse(ar.Request.UID, org, obj)
					log.Infof("%s[cost:%v] do review done", logPrefix, cost)
					return ar
				},
			},
		},
	}
}
