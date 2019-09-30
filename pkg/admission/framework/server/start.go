package server

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/model"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) startModels() (err error) {
	// filter
	filter := util.MakeModelEnabledFilter(s.enableOptions)
	allModels := s.modelManager.ListModels()
	models := make([]model.Model, 0, len(allModels))
	for _, m := range allModels {
		if filter(m.Name()) {
			models = append(models, m)
		} else {
			log.Warningf("model %s filter out for not enabled", m.Name())
		}
	}
	// start
	go s.informerFactory.Start(s.stopCh)
	for _, m := range models {
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
