package v1

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	deploymentGRV = appsv1.SchemeGroupVersion.WithResource("deployment")
	deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
)

type DeploymentProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.Deployment) (err error)
}

type DeploymentConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map Deployment processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]DeploymentProcessor
}
