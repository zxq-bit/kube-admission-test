package server

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/model"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/models/demo"
)

type ModelMaker func() (model.Model, error)

func (s *Server) ensureModelMaker() {
	// demo
	s.modelManager.RegisterMaker(
		demo.ModelName,
		func() (model.Model, error) {
			return demo.NewModel(s.kc, s.informerFactory)
		},
	)
}
