package server

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/module"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) startModules() (err error) {
	// filter
	filter := util.MakeModuleEnabledFilter(s.enableOptions)
	allModules := s.moduleManager.ListModules()
	modules := make([]module.Module, 0, len(allModules))
	for _, m := range allModules {
		if filter(m.Name()) {
			modules = append(modules, m)
		} else {
			log.Warningf("module %s filter out for not enabled", m.Name())
		}
	}
	// start
	go s.informerFactory.Start(s.stopCh)
	for _, m := range modules {
		go m.Start(s.stopCh)
	}
	// synced
	synced := s.informerFactory.WaitForCacheSync(s.stopCh)
	for tp, isSync := range synced {
		tpStr := fmt.Sprintf("%s.%s", tp.PkgPath(), tp.Name())
		if isSync {
			log.Infof("informer for %v synced", tpStr)
		} else {
			msg := fmt.Sprintf("informer for %v not synced", tpStr)
			err = fmt.Errorf(msg)
			log.Errorf(msg)
		}
	}
	return err
}
