package v1

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	persistentvolumeGRV = corev1.SchemeGroupVersion.WithResource("persistentvolume")
	persistentvolumeGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolume")
)

type PersistentVolumeProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *corev1.PersistentVolume) (err error)
}

type PersistentVolumeConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map PersistentVolume processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PersistentVolumeProcessor
}
