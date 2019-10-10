package server

import (
	"fmt"
	"path"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) initModules() error {
	moduleFilter := util.MakeModuleEnabledFilter(s.enableOptions)
	e := s.moduleManager.ExecuteMakers(moduleFilter)
	if e != nil {
		log.Errorf("initModules ExecuteMakers failed, %v", e)
	}
	return e
}

func (s *Server) initReviews() error {
	moduleFilter := util.MakeModuleEnabledFilter(s.enableOptions)
	for _, h := range s.handlerConfig.Handlers {
		gvr := h.GroupVersionResource
		opType := h.OpType
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
			module := s.moduleManager.GetModule(pMeta.Module)
			if module == nil {
				log.Errorf("%s module not found", logPrefix)
				return fmt.Errorf("module %s not register in server", pMeta.Module)
			}
			p := module.GetProcessor(pMeta.Name)
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
	return nil
}
