package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	statefulsetsGVR = appsv1.SchemeGroupVersion.WithResource("statefulsets")
	statefulsetsGVK = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
)

type StatefulSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	processor.Tracer
	// Review do review, return error if should stop
	Review func(in *appsv1.StatefulSet) (err error)
}

type StatefulSetConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map StatefulSet processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]*StatefulSetProcessor
}
