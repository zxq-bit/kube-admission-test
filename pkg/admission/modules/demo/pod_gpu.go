package demo

import (
	"context"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	rcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/apis/core/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	envKeyNvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"
	resourceKeyNvGPU           = "nvidia.com/gpu"
	resourceKeyPrefixER        = "extendedresource.caicloud.io/"
	resourceKeyPrefixERReq     = "requests.extendedresource.caicloud.io/"
)

var (
	podProcessorGPUVisible = &rcorev1.PodProcessor{
		Metadata: processor.Metadata{
			Name:       ProcessorNamePodGPUVisible,
			ModuleName: ModuleName,
			IgnoreNamespaces: []string{
				metav1.NamespaceDefault,
				metav1.NamespaceSystem,
			},
			Type: constants.ProcessorTypeMutate,
		},
		Review: func(ctx context.Context, in *corev1.Pod) (err error) {
			mutatePodForGPUEnv(in)
			return
		},
	}
)

func mutatePodForGPUEnv(in *corev1.Pod) {
	containerSets := [][]corev1.Container{
		in.Spec.InitContainers,
		in.Spec.Containers,
	}
	for _, containers := range containerSets {
		for i := range containers {
			mutateContainerForGPUEnv(&in.Spec.Containers[i])
		}
	}
}

func mutateContainerForGPUEnv(in *corev1.Container) {
	// using gpu
	if isResourceRequirementsContainsGPU(&in.Resources) {
		return
	}
	// set env
	found := false
	for i := range in.Env {
		env := &in.Env[i]
		if env.Name == envKeyNvidiaVisibleDevices {
			env.Value = ""
			found = true
		}
	}
	if !found {
		in.Env = append(in.Env, corev1.EnvVar{
			Name:  envKeyNvidiaVisibleDevices,
			Value: "",
		})
	}
}

func isResourceRequirementsContainsGPU(r *corev1.ResourceRequirements) bool {
	lists := []corev1.ResourceList{
		r.Requests,
		r.Limits,
	}
	for _, list := range lists {
		for k, q := range list {
			// resource class
			if (k == resourceKeyNvGPU ||
				strings.HasPrefix(string(k), resourceKeyPrefixER) ||
				strings.HasPrefix(string(k), resourceKeyPrefixERReq)) &&
				q.Value() > 0 {
				return true
			}
		}
	}
	return false
}
