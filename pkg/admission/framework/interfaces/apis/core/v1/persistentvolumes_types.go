package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	persistentvolumesGRV = corev1.SchemeGroupVersion.WithResource("persistentvolumes")
	persistentvolumesGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolume")
)

type PersistentVolumeProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.PersistentVolume) (err error)
}

type PersistentVolumeConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map PersistentVolume processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PersistentVolumeProcessor
}
