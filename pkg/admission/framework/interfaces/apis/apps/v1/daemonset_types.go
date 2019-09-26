package v1

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	daemonsetGRV = appsv1.SchemeGroupVersion.WithResource("daemonset")
	daemonsetGVK = appsv1.SchemeGroupVersion.WithKind("DaemonSet")
)

type DaemonSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.DaemonSet) (err error)
}

type DaemonSetConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map DaemonSet processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]DaemonSetProcessor
}
