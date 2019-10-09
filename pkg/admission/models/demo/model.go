package demo

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/model"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/go-common/interfaces"

	listerscorev1 "k8s.io/client-go/listers/core/v1"
)

type Model struct {
	kc kubernetes.Interface
	f  informers.SharedInformerFactory

	configMapLister listerscorev1.ConfigMapLister
	secretLister    listerscorev1.SecretLister

	pMap map[string]interface{}
}

const (
	ModelName = "demo"

	ProcessorNameCmExample     = "ConfigMapExample"
	ProcessorNamePodExample    = "PodExample"
	ProcessorNamePodGPUVisible = "PodGPUVisible"
	ProcessorNameDpCheckMntRef = "DpCheckMntRef"
)

func NewModel(kc kubernetes.Interface, f informers.SharedInformerFactory) (model.Model, error) {
	if interfaces.IsNil(kc) {
		return nil, fmt.Errorf("nil kubernetes client")
	}
	if interfaces.IsNil(f) {
		return nil, fmt.Errorf("nil kubernetes informer factory")
	}
	m := &Model{
		kc: kc,
		f:  f,
	}

	m.configMapLister = f.Core().V1().ConfigMaps().Lister()
	m.secretLister = f.Core().V1().Secrets().Lister()

	m.pMap = map[string]interface{}{
		ProcessorNameCmExample:     cmProcessorExample,
		ProcessorNamePodExample:    podProcessorExample,
		ProcessorNamePodGPUVisible: podProcessorGPUVisible,
		ProcessorNameDpCheckMntRef: m.getDpProcessorCheckMntRef(),
	}
	return m, nil
}

func (m *Model) Name() string { return ModelName }

func (m *Model) Start(stopCh <-chan struct{}) {}

func (m *Model) GetProcessor(name string) interface{} {
	if p, ok := m.pMap[name]; ok && p != nil {
		return p
	}
	return nil
}
