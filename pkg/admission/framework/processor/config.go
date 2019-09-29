package processor

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/logger"

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/go-common/nirvana/middleware"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RawExtensionParser func(raw *runtime.RawExtension) (runtime.Object, error)

type Config struct {
	// Group+Version+Resource
	GroupVersionResource schema.GroupVersionResource
	// RawExtensionParser parser of RawExtension
	RawExtensionParser RawExtensionParser
	// TimeoutSecondsMap set total execute time of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map processors by operation type
	ProcessorsMap map[arv1b1.OperationType]func(ctx context.Context, in runtime.Object) (err error)
}

var (
	// failurePolicyTypeIgnore = arv1b1.Ignore
	failurePolicyTypeFail = arv1b1.Fail
)

func (c *Config) ToMutatingWebHook(opt *StartOptions) (re *arv1b1.MutatingWebhookConfiguration) {
	gvr := c.GroupVersionResource
	re = &arv1b1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.JoinObjectName(gvr.Group, gvr.Version, gvr.Resource),
		},
		Webhooks: make([]arv1b1.Webhook, 0, len(c.ProcessorsMap)),
	}
	for _, opType := range constants.OperationTypes {
		if c.ProcessorsMap[opType] == nil {
			continue
		}
		var timeoutSeconds *int32
		if seconds, ok := c.TimeoutSecondsMap[opType]; ok && seconds > 0 {
			timeoutSeconds = &seconds
		}
		re.Webhooks = append(re.Webhooks, arv1b1.Webhook{
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
		})
	}
	if len(re.Webhooks) == 0 {
		return nil
	}
	return re
}

func (c *Config) ToNirvanaDescriptors(opt *StartOptions) (re []definition.Descriptor) {
	if len(c.ProcessorsMap) == 0 {
		return
	}
	gvr := c.GroupVersionResource
	for _, opType := range constants.OperationTypes {
		f := c.ProcessorsMap[opType]
		if f == nil {
			continue
		}
		// timeout
		middlewares := []definition.Middleware{
			logger.New(log.DefaultLogger()),
		}
		if seconds, ok := c.TimeoutSecondsMap[opType]; ok && seconds > 0 {
			middlewares = append(middlewares, middleware.Timeout(time.Duration(seconds)*time.Second))
		}
		logBase := fmt.Sprintf("[%v/%v/%v|%v]", gvr.Group, gvr.Version, gvr.Resource, opType)
		// descriptor
		re = append(re, definition.Descriptor{
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
						org, e := c.RawExtensionParser(&ar.Request.Object)
						if e != nil {
							log.Errorf("%s RawExtensionParser failed, %v", logPrefix, e)
							ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
							return ar
						}
						log.DefaultLogger().Infof("%s RawExtensionParser done", logPrefix)
						// do review
						obj := org.DeepCopyObject()
						if e = f(ctx, obj); e != nil {
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
		})
	}
	return
}

func joinWebHooksPath(rootPath string, gvk schema.GroupVersionResource,
	opType arv1b1.OperationType) *string {
	result := path.Join(rootPath, gvk.Group, gvk.Version, gvk.Resource, strings.ToLower(string(opType)))
	return &result
}
