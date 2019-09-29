package v1

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	deploymentsGVR = appsv1.SchemeGroupVersion.WithResource("deployments")
	deploymentsGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
)

type DeploymentProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	processor.Tracer
	// Review do review, return error if should stop
	Review func(in *appsv1.Deployment) (err error)
}

type DeploymentConfig struct {
	// TimeoutSecondsMap set total execute time by second of processors
	TimeoutSecondsMap map[arv1b1.OperationType]int32
	// ProcessorsMap map Deployment processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]*DeploymentProcessor
}
