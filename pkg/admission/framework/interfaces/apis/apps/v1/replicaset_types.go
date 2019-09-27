package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	replicasetGRV = appsv1.SchemeGroupVersion.WithResource("replicaset")
	replicasetGVK = appsv1.SchemeGroupVersion.WithKind("ReplicaSet")
)

type ReplicaSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.ReplicaSet) (err error)
}

type ReplicaSetConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map ReplicaSet processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]ReplicaSetProcessor
}
