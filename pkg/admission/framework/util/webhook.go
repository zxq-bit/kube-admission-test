package util

import (
	"path"
	"strings"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// failurePolicyTypeIgnore = arv1b1.Ignore
	failurePolicyTypeFail = arv1b1.Fail
)

func MakeMutatingWebHooks(svcConfig *ServiceConfig, gvk metav1.GroupVersionResource,
	opType arv1b1.OperationType) arv1b1.MutatingWebhookConfiguration {
	return arv1b1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workload-mutating",
		},
		Webhooks: []arv1b1.Webhook{
			{
				Name: makeMutatingWebHooksName(gvk.Group, gvk.Version, gvk.Resource, string(opType)),
				Rules: []arv1b1.RuleWithOperations{
					{
						Operations: []arv1b1.OperationType{opType},
						Rule: arv1b1.Rule{
							APIGroups:   []string{gvk.Group},
							APIVersions: []string{gvk.Version},
							Resources:   []string{gvk.Resource},
						},
					},
				},
				ClientConfig: arv1b1.WebhookClientConfig{
					Service: &arv1b1.ServiceReference{
						Name:      svcConfig.Name,
						Namespace: svcConfig.Namespace,
						Path:      makeMutatingWebHooksPath(svcConfig.RootPath, gvk, opType),
					},
					CABundle: svcConfig.CABundle,
				},
				FailurePolicy: &failurePolicyTypeFail,
			},
		},
	}
}

func makeMutatingWebHooksName(array ...string) string {
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

func makeMutatingWebHooksPath(rootPath string, gvk metav1.GroupVersionResource,
	opType arv1b1.OperationType) *string {
	result := path.Join(rootPath, gvk.Group, gvk.Version, gvk.Resource, strings.ToLower(string(opType)))
	return &result
}
