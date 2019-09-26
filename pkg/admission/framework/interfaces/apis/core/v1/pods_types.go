package v1

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	podsGRV = corev1.SchemeGroupVersion.WithResource("pods")
	podsGVK = corev1.SchemeGroupVersion.WithKind("Pod")
)

type PodProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.Pod) (err error)
}

type PodConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map Pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PodProcessor
}
