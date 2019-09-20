package v1alpha1

import "fmt"

type WorkloadType string

const (
	DaemonSetWorkloadType = "daemonset"

	WorkloadAnnotationCreateBy    = "workload.caicloud.io/createType"
	WorkloadAnnotationAlias       = "workload.caicloud.io/alias"
	WorkloadAnnotationDescription = "workload.caicloud.io/description"
	HelmAnnotationName            = "helm.sh/release"
	HelmAnnotationNamespace       = "helm.sh/namespace"
	HelmAnnotationPath            = "helm.sh/path"

	WorkloadTaintLabelsName = "controller.caicloud.io/release"
	WorkloadTaintLabelsKind = "controller.caicloud.io/kind"

	WorkloadAdmissionServiceName = "workload-admission"

	WorkloadLabelReactor       = "workload.caicloud.io/reactor"
	WorkloadLabelKind          = "workload.caicloud.io/kind"
	WorkloadLabelLastOperation = "workload.caicloud.io/lastOperation"

	WorkloadLabelOperationSuspend = "suspend"

	WorkNameSeparator = "-"

	WorkloadDaemonSet WorkloadType = "daemonset"
)

// GetWorkloadKindAndName gets workload's kind and name
func GetWorkloadKindAndName(labels map[string]string) (string, string, error) {
	kind, ok := labels[WorkloadLabelKind]
	if !ok {
		return "", "", fmt.Errorf("labels find kind %s failed", WorkloadLabelKind)
	}
	name, ok := labels[WorkloadLabelReactor]
	if !ok {
		return "", "", fmt.Errorf("labels find name %s failed", WorkloadLabelReactor)
	}
	return kind, name, nil
}
