package processor

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/logger"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/go-common/nirvana/middleware"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Review struct {
	// TimeoutSecond set total execute time of processors
	TimeoutSecond int32
	// Tracer do trace and save data
	Tracer Tracer
	// Review
	Review func(ctx context.Context, in runtime.Object) (err error)
}

func (r *Review) ToWebHook(opt *StartOptions, gvr schema.GroupVersionResource, opType arv1b1.OperationType) *arv1b1.Webhook {
	var timeoutSeconds *int32
	if r.TimeoutSecond > 0 {
		timeoutSeconds = &r.TimeoutSecond
	}
	webhook := &arv1b1.Webhook{
		Name: util.JoinObjectName(gvr.Group, gvr.Version, gvr.Resource, string(opType)),
		Rules: []arv1b1.RuleWithOperations{
			{
				Operations: []arv1b1.OperationType{opType},
				Rule: arv1b1.Rule{
					APIGroups:   []string{gvr.Group},
					APIVersions: []string{gvr.Version},
					Resources:   []string{gvr.Resource},
				},
			},
		},
		ClientConfig: arv1b1.WebhookClientConfig{
			Service: &arv1b1.ServiceReference{
				Name:      opt.ServiceName,
				Namespace: opt.ServiceNamespace,
				Path:      joinWebHooksPath(opt.APIRootPath, gvr, opType),
			},
			CABundle: opt.ServiceCABundle,
		},
		TimeoutSeconds: timeoutSeconds,
		FailurePolicy:  &failurePolicyTypeFail,
	}
	return webhook
}

func (r *Review) ToNirvanaDescriptors(opt *StartOptions,
	gvr schema.GroupVersionResource, opType arv1b1.OperationType, rawExtensionParser RawExtensionParser) definition.Descriptor {
	// timeout
	middlewares := []definition.Middleware{
		logger.New(log.DefaultLogger()),
		tracer.New(&r.Tracer),
	}
	if r.TimeoutSecond > 0 {
		middlewares = append(middlewares, middleware.Timeout(time.Duration(r.TimeoutSecond)*time.Second))
	}
	logBase := fmt.Sprintf("[%v/%v/%v|%v]", gvr.Group, gvr.Version, gvr.Resource, opType)
	// descriptor
	descriptor := definition.Descriptor{
		Path:        *joinWebHooksPath(opt.APIRootPath, gvr, opType),
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
					org, e := rawExtensionParser(&ar.Request.Object)
					if e != nil {
						log.Errorf("%s RawExtensionParser failed, %v", logPrefix, e)
						ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
						return ar
					}
					log.DefaultLogger().Infof("%s RawExtensionParser done", logPrefix)
					// do review
					obj := org.DeepCopyObject()
					if e = r.Review(ctx, obj); e != nil {
						log.Errorf("%s do review failed, %v", logPrefix, e)
						ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
						return ar
					}
					log.Infof("%s do review done", logPrefix)
					ar.Response = util.ToAdmissionPassResponse(ar.Request.UID, org, obj)
					return ar
				},
			},
		},
	}
	return descriptor
}
