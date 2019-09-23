package util

import (
	"fmt"
	"path"
	"strings"

	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// failurePolicyTypeIgnore = arv1b1.Ignore
	failurePolicyTypeFail = arv1b1.Fail
)

func MakeMutatingWebHook(svcConfig *ServiceConfig, gvr schema.GroupVersionResource,
	opType arv1b1.OperationType) arv1b1.Webhook {
	return arv1b1.Webhook{
		Name: makeMutatingWebHookName(gvr.Group, gvr.Version, gvr.Resource, string(opType)),
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
				Name:      svcConfig.Name,
				Namespace: svcConfig.Namespace,
				Path:      makeMutatingWebHooksPath(svcConfig.RootPath, gvr, opType),
			},
			CABundle: svcConfig.CABundle,
		},
		FailurePolicy: &failurePolicyTypeFail,
	}
}

func MakeNirvanaDescriptor(svcConfig *ServiceConfig, gvr schema.GroupVersionResource,
	opType arv1b1.OperationType, f ReviewFuncWithContext) definition.Descriptor {
	return definition.Descriptor{
		Path: *makeMutatingWebHooksPath(svcConfig.RootPath, gvr, opType),
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Description: fmt.Sprintf("do admission for %v/%v/%v.%v", gvr.Group, gvr.Version, gvr.Resource, opType),
				Parameters: []definition.Parameter{
					{
						Source:      definition.Body,
						Name:        "body",
						Description: "admission body",
					},
				},
				Results:  definition.DataErrorResults("admission response"),
				Function: f,
			},
		},
	}
}

func makeMutatingWebHookName(array ...string) string {
	emptyCount := 0
	for _, s := range array {
		if s == "" {
			emptyCount++
		}
	}
	vec := make([]string, 0, len(array)-emptyCount)
	for _, s := range array {
		if s == "" {
			continue
		}
		vec = append(vec, s)
	}
	return strings.ToLower(strings.Join(vec, "."))
}

func makeMutatingWebHooksPath(rootPath string, gvk schema.GroupVersionResource,
	opType arv1b1.OperationType) *string {
	result := path.Join(rootPath, gvk.Group, gvk.Version, gvk.Resource, strings.ToLower(string(opType)))
	return &result
}
