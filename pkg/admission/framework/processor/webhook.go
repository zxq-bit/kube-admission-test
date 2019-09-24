package processor

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/options"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/nirvana/definition"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// failurePolicyTypeIgnore = arv1b1.Ignore
	failurePolicyTypeFail = arv1b1.Fail
)

func makeMutatingWebHook(opt *options.StartOptions, gvr schema.GroupVersionResource,
	opType arv1b1.OperationType) arv1b1.Webhook {
	return arv1b1.Webhook{
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
	}
}

func makeNirvanaDescriptor(apiRootPath string, gvr schema.GroupVersionResource,
	opType arv1b1.OperationType, f util.ReviewFuncWithContext) definition.Descriptor {
	arGVK := admissionv1b1.SchemeGroupVersion.WithKind(reflect.TypeOf(admissionv1b1.AdmissionReview{}).Name())
	return definition.Descriptor{
		Path: *joinWebHooksPath(apiRootPath, gvr, opType),
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
									ar := &admissionv1b1.AdmissionReview{}
									deserializer := scheme.Codecs.UniversalDeserializer()
									if _, _, err := deserializer.Decode(body, &arGVK, ar); err != nil {
										ar.Response = util.ToAdmissionFailedResponse("", err)
									}
									return ar, nil
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
					objCopy := ar.Request.Object.Object.DeepCopyObject()
					if e := f(ctx, objCopy); e != nil {
						ar.Response = util.ToAdmissionFailedResponse(ar.Request.UID, e)
						return ar
					}
					ar.Response = util.ToAdmissionPassResponse(ar.Request.UID, )
				},
			},
		},
	}
}

func joinWebHooksPath(rootPath string, gvk schema.GroupVersionResource,
	opType arv1b1.OperationType) *string {
	result := path.Join(rootPath, gvk.Group, gvk.Version, gvk.Resource, strings.ToLower(string(opType)))
	return &result
}
