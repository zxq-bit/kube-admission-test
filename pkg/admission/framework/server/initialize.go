package server

import (
	"fmt"
	"path"

	// call registers on pkg init
	_ "github.com/zxq-bit/kube-admission-test/pkg/admission/install"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/module"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) initModules() (err error) {
	moduleFilter := util.MakeModuleEnabledFilter(s.enableOptions)
	moduleMaker := module.GetModuleMakerManager()
	s.moduleManager, err = moduleMaker.ExecuteMakers(s.kc, s.informerFactory, moduleFilter)
	if err != nil {
		log.Errorf("initModules ExecuteMakers failed, %v", err)
	}
	return err
}

func (s *Server) initReviews() error {
	moduleFilter := util.MakeModuleEnabledFilter(s.enableOptions)
	for _, gvrHandler := range s.handlerConfig.Configs {
		// by GVR
		gvr := gvrHandler.GroupVersionResource
		for opType, h := range gvrHandler.Handlers {
			logBase := fmt.Sprintf("handlerInit[%s][%v]", path.Join(gvr.Group, gvr.Version, gvr.Resource), opType)
			// get handler
			handler, e := s.reviewManager.GetHandler(gvr, opType)
			if e != nil {
				log.Errorf("%s GetHandler failed, %v", logBase, e)
				return e
			}
			// set timeout
			e = handler.SetTimeout(h.TimeoutSecond)
			if e != nil {
				log.Errorf("%s SetTimeout failed, %v", logBase, e)
				return e
			}
			// register processors
			for i, pMeta := range h.Processors {
				logPrefix := fmt.Sprintf("%s[%d][%s/%s]", logBase, i, pMeta.Module, pMeta.Name)
				// filter by module
				if !moduleFilter(pMeta.Module) {
					log.Warningf("%s skipped for not enabled", logPrefix)
					continue
				}
				// get module processor
				m := s.moduleManager.GetModule(pMeta.Module)
				if m == nil {
					log.Errorf("%s module not found", logPrefix)
					return fmt.Errorf("module %s not register in server", pMeta.Module)
				}
				p := m.GetProcessor(pMeta.Name)
				if p == nil {
					log.Errorf("%s processor not found", logPrefix)
					return fmt.Errorf("porcessor %s not register in module %s", pMeta.Name, pMeta.Module)
				}
				// register
				e = handler.Register(p)
				if e != nil {
					log.Errorf("%s processor Register failed, %v", logPrefix, e)
					return e
				}
				log.Infof("%s processor Register done", logPrefix)
			}
		}
	}
	return nil
}
