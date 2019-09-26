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

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/go-common/nirvana/middleware"
	"github.com/caicloud/nirvana/definition"

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
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map processors by operation type
	ProcessorsMap map[arv1b1.OperationType]util.Review
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
			FailurePolicy: &failurePolicyTypeFail,
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
		var middlewares []definition.Middleware
		if timeout, ok := c.TimeoutMap[opType]; ok && timeout > 0 {
			middlewares = append(middlewares, middleware.Timeout(timeout))
		}
		// descriptor
		re = append(re, definition.Descriptor{
			Path:        *joinWebHooksPath(opt.APIRootPath, gvr, opType),
			Middlewares: middlewares,
			Definitions: []definition.Definition{
				{
					Method:      definition.Create,
					Description: fmt.Sprintf("do admission for %v/%v/%v.%v", gvr.Group, gvr.Version, gvr.Resource, opType),
					Parameters: []definition.Parameter{
						{
							Source:      definition.Body,
							Name:        "admissionReview",
							Description: "admission body",
							Operators: []definition.Operator{
								definition.OperatorFunc("auto scaling group filter",
									func(ctx context.Context, name string, body []byte) (*admissionv1b1.AdmissionReview, error) {
										ar := admissionv1b1.AdmissionReview{}
										gvk := admissionv1b1.SchemeGroupVersion.WithKind(reflect.TypeOf(ar).Name())
										deserializer := scheme.Codecs.UniversalDeserializer()
										if _, _, err := deserializer.Decode(body, &gvk, &ar); err != nil {
											ar.Response = util.ToAdmissionFailedResponse("", err)
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
						if ar.Response != nil {
							return ar
						}
						org, e := c.RawExtensionParser(&ar.Request.Object)
						if e != nil {
							ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
							return ar
						}
						obj := org.DeepCopyObject()
						if e = f(ctx, obj); e != nil {
							ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
							return ar
						}
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
