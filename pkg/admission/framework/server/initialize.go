package server

import (
	"fmt"
	"path"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"
)

func (s *Server) initModels() error {
	modelFilter := util.MakeModelEnabledFilter(s.enableOptions)
	e := s.modelManager.ExecuteMakers(modelFilter)
	if e != nil {
		log.Errorf("initModels ExecuteMakers failed, %v", e)
	}
	return e
}

func (s *Server) initReviews() error {
	modelFilter := util.MakeModelEnabledFilter(s.enableOptions)
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
			logPrefix := fmt.Sprintf("%s[%d][%s/%s]", logBase, i, pMeta.Model, pMeta.Name)
			// filter by model
			if !modelFilter(pMeta.Model) {
				log.Warningf("%s skipped for not enabled", logPrefix)
				continue
			}
			// get model processor
			model := s.modelManager.GetModel(pMeta.Model)
			if model == nil {
				log.Errorf("%s model not found", logPrefix)
				return fmt.Errorf("model %s not register in server", pMeta.Model)
			}
			p := model.GetProcessor(pMeta.Name)
			if p == nil {
				log.Errorf("%s processor not found", logPrefix)
				return fmt.Errorf("porcessor %s not register in model %s", pMeta.Name, pMeta.Model)
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
