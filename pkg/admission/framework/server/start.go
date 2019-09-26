package server

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) startModels(opt processor.StartOptions) (err error) {
	// filter
	filter := util.MakeModelEnabledFilter(opt.EnableOptions)
	allModels := s.modelCollection.ListModels()
	models := make([]interfaces.Model, 0, len(allModels))
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
		if !isSync {
			msg := fmt.Sprintf("informer for %v not synced", tp.Name())
			err = fmt.Errorf(msg)
			log.Errorf(msg)
		}
	}
	return err
}
