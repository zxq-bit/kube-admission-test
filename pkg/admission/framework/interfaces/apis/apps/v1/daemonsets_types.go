package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	daemonsetsGVR = appsv1.SchemeGroupVersion.WithResource("daemonsets")
	daemonsetsGVK = appsv1.SchemeGroupVersion.WithKind("DaemonSet")
)

type DaemonSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	processor.Tracer
	// Review do review, return error if should stop
	Review func(in *appsv1.DaemonSet) (err error)
}

type DaemonSetConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map DaemonSet processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]*DaemonSetProcessor
}
