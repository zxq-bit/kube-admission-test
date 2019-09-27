package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	configmapGRV = corev1.SchemeGroupVersion.WithResource("configmap")
	configmapGVK = corev1.SchemeGroupVersion.WithKind("ConfigMap")
)

type ConfigMapProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.ConfigMap) (err error)
}

type ConfigMapConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map ConfigMap processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]ConfigMapProcessor
}
