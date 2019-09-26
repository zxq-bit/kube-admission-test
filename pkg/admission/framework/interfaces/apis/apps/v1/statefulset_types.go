package v1

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	statefulsetGRV = appsv1.SchemeGroupVersion.WithResource("statefulset")
	statefulsetGVK = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
)

type StatefulSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.StatefulSet) (err error)
}

type StatefulSetConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map StatefulSet processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]StatefulSetProcessor
}
