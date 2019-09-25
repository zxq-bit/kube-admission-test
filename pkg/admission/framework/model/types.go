package model

import (
	"fmt"
	"sort"

	"github.com/caicloud/go-common/interfaces"
)

type Model interface {
	Name() string
	Start(stopCh <-chan struct{})
}

type Collection struct {
	ModelMap map[string]Model
}

func (c *Collection) Register(models ...Model) error {
	for _, m := range models {
		if interfaces.IsNil(m) {
			return fmt.Errorf("model is nil")
		}
		if c.ModelMap == nil {
			c.ModelMap = make(map[string]Model, 1)
		}
		if _, exist := c.ModelMap[m.Name()]; exist {
			return fmt.Errorf("model %s already exist", m.Name())
		}
		c.ModelMap[m.Name()] = m
	}
	return nil
}

func (c *Collection) ListModels() []Model {
	l := len(c.ModelMap)
	if l == 0 {
		return nil
	}
	re := make([]Model, 0, l)
	for _, m := range c.ModelMap {
		re = append(re, m)
	}
	sort.Slice(re, func(i, j int) bool {
		return re[i].Name() < re[j].Name()
	})
	return re
}
