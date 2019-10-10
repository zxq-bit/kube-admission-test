package module

import (
	"fmt"
	"sort"
	"sync"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"
)

var (
	moduleMakerManager MakerManager
)

func GetModuleMakerManager() *MakerManager {
	return &moduleMakerManager
}

func RegisterMaker(key string, maker Maker) {
	moduleMakerManager.Register(key, maker)
}

type Maker func(kc kubernetes.Interface, f informers.SharedInformerFactory) (Module, error)

type MakerManager struct {
	makerMap map[string]Maker
	lock     sync.RWMutex
}

type Manager struct {
	moduleMap map[string]Module
}

func (c *MakerManager) Register(key string, maker Maker) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.makerMap == nil {
		c.makerMap = map[string]Maker{}
	}
	c.makerMap[key] = maker

	log.Infof("ModuleMaker register: %s", key)
}

func (c *MakerManager) ExecuteMakers(kc kubernetes.Interface, f informers.SharedInformerFactory,
	filter util.ModuleEnabledFilter) (re Manager, e error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.makerMap) == 0 {
		return re, nil
	}
	if filter == nil {
		filter = func(name string) bool { return true }
	}
	for name, maker := range c.makerMap {
		logPrefix := fmt.Sprintf("executeModuleMaker[%s]", name)
		// filter
		if !filter(name) {
			log.Warning("%s skipped for not enabled", logPrefix)
			continue
		}
		// do make
		module, e := maker(kc, f)
		if e != nil {
			log.Errorf("%s execute failed, %v", logPrefix, e)
			return re, e
		}
		// register
		e = re.registerModules(module)
		if e != nil {
			log.Errorf("%s register failed, %v", logPrefix, e)
			return re, e
		}
	}
	return re, nil
}

func (c *Manager) registerModules(modules ...Module) error {
	for _, m := range modules {
		if interfaces.IsNil(m) {
			return fmt.Errorf("module is nil")
		}
		if c.moduleMap == nil {
			c.moduleMap = make(map[string]Module, 1)
		}
		if _, exist := c.moduleMap[m.Name()]; exist {
			return fmt.Errorf("module %s already exist", m.Name())
		}
		c.moduleMap[m.Name()] = m
	}
	return nil
}

func (c *Manager) GetModule(name string) Module {
	if c.moduleMap == nil {
		return nil
	}
	return c.moduleMap[name]
}

func (c *Manager) ListModules() []Module {
	l := len(c.moduleMap)
	if l == 0 {
		return nil
	}
	re := make([]Module, 0, l)
	for _, m := range c.moduleMap {
		re = append(re, m)
	}
	sort.Slice(re, func(i, j int) bool {
		return re[i].Name() < re[j].Name()
	})
	return re
}
