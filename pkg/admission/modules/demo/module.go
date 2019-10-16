package demo

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/module"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/go-common/interfaces"

	listerscorev1 "k8s.io/client-go/listers/core/v1"
)

func init() {
	module.RegisterMaker(ModuleName, NewModule)
}

type Module struct {
	kc kubernetes.Interface
	f  informers.SharedInformerFactory

	configMapLister listerscorev1.ConfigMapLister
	secretLister    listerscorev1.SecretLister

	pMap map[string]interface{}
}

const (
	ModuleName = "demo"

	ProcessorNameCmExample       = "ConfigMapExample"
	ProcessorNameCmDeletionAllow = "ConfigMapDeletionAllow"
	ProcessorNameCmPanicPass     = "ConfigMapPanicPass"
	ProcessorNamePodExample      = "PodExample"
	ProcessorNamePodGPUVisible   = "PodGPUVisible"
	ProcessorNameDpCheckMntRef   = "DpCheckMntRef"
)

func NewModule(kc kubernetes.Interface, f informers.SharedInformerFactory) (module.Module, error) {
	if interfaces.IsNil(kc) {
		return nil, fmt.Errorf("nil kubernetes client")
	}
	if interfaces.IsNil(f) {
		return nil, fmt.Errorf("nil kubernetes informer factory")
	}
	m := &Module{
		kc: kc,
		f:  f,
	}

	m.configMapLister = f.Core().V1().ConfigMaps().Lister()
	m.secretLister = f.Core().V1().Secrets().Lister()

	m.pMap = map[string]interface{}{
		ProcessorNameCmExample:       cmProcessorExample,
		ProcessorNameCmDeletionAllow: cmProcessorDeletionAllow,
		ProcessorNameCmPanicPass:     cmProcessorPanicPass,
		ProcessorNamePodExample:      podProcessorExample,
		ProcessorNamePodGPUVisible:   podProcessorGPUVisible,
		ProcessorNameDpCheckMntRef:   m.getDpProcessorCheckMntRef(),
	}
	return m, nil
}

func (m *Module) Name() string { return ModuleName }

func (m *Module) Start(stopCh <-chan struct{}) {}

func (m *Module) GetProcessor(name string) interface{} {
	if p, ok := m.pMap[name]; ok && p != nil {
		return p
	}
	return nil
}
