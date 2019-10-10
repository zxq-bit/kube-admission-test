package module

import (
	"fmt"
	"sort"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"
)

type Maker func() (Module, error)

type Manager struct {
	makerMap  map[string]Maker
	moduleMap map[string]Module
}

func (c *Manager) RegisterMaker(key string, maker Maker) {
	if c.makerMap == nil {
		c.makerMap = map[string]Maker{}
	}
	c.makerMap[key] = maker
}

func (c *Manager) ExecuteMakers(filter util.ModuleEnabledFilter) error {
	if len(c.makerMap) == 0 {
		return nil
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
		module, e := maker()
		if e != nil {
			log.Errorf("%s execute failed, %v", logPrefix, e)
			return e
		}
		// register
		e = c.registerModules(module)
		if e != nil {
			log.Errorf("%s register failed, %v", logPrefix, e)
			return e
		}
	}
	return nil
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
