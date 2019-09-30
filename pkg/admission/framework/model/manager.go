package model

import (
	"fmt"
	"sort"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

type Maker func() (Model, error)

type Manager struct {
	makerMap map[string]Maker
	modelMap map[string]Model
}

func (c *Manager) RegisterMaker(key string, maker Maker) {
	if c.makerMap == nil {
		c.makerMap = map[string]Maker{}
	}
	c.makerMap[key] = maker
}

func (c *Manager) ExecuteMakers(filter util.ModelEnabledFilter) error {
	if len(c.makerMap) == 0 {
		return nil
	}
	if filter == nil {
		filter = func(name string) bool { return true }
	}
	for name, maker := range c.makerMap {
		logPrefix := fmt.Sprintf("executeModelMaker[%s]", name)
		// filter
		if !filter(name) {
			log.Warning("%s skipped for not enabled", logPrefix)
			continue
		}
		// do make
		model, e := maker()
		if e != nil {
			log.Errorf("%s execute failed, %v", logPrefix, e)
			return e
		}
		// register
		e = c.registerModels(model)
		if e != nil {
			log.Errorf("%s register failed, %v", logPrefix, e)
			return e
		}
	}
	return nil
}

func (c *Manager) registerModels(models ...Model) error {
	for _, m := range models {
		if interfaces.IsNil(m) {
			return fmt.Errorf("model is nil")
		}
		if c.modelMap == nil {
			c.modelMap = make(map[string]Model, 1)
		}
		if _, exist := c.modelMap[m.Name()]; exist {
			return fmt.Errorf("model %s already exist", m.Name())
		}
		c.modelMap[m.Name()] = m
	}
	return nil
}

func (c *Manager) GetModel(name string) Model {
	if c.modelMap == nil {
		return nil
	}
	return c.modelMap[name]
}

func (c *Manager) ListModels() []Model {
	l := len(c.modelMap)
	if l == 0 {
		return nil
	}
	re := make([]Model, 0, l)
	for _, m := range c.modelMap {
		re = append(re, m)
	}
	sort.Slice(re, func(i, j int) bool {
		return re[i].Name() < re[j].Name()
	})
	return re
}
