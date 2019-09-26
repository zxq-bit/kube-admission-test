package app

import (
	"fmt"

	acorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces/apis/core/v1"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/go-common/interfaces"
)

type Model struct {
	kc kubernetes.Interface
	f  informers.SharedInformerFactory

	podMap map[string]*acorev1.PodProcessor
}

func NewModel(kc kubernetes.Interface, f informers.SharedInformerFactory) (*Model, error) {
	if interfaces.IsNil(kc) {
		return nil, fmt.Errorf("nil kubernetes client")
	}
	if interfaces.IsNil(f) {
		return nil, fmt.Errorf("nil kubernetes informer factory")
	}
	m := &Model{
		kc: kc,
		f:  f,
		podMap: map[string]*acorev1.PodProcessor{
			ProcessorNamePodGPUVisible: podProcessorGPUVisible,
		},
	}
	return m, nil
}

func (m *Model) Name() string { return ModelName }

func (m *Model) Start(stopCh <-chan struct{}) {}

func (m *Model) GetPodProcessor(name string) *acorev1.PodProcessor {
	return m.podMap[name]
}
