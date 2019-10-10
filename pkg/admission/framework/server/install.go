package server

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/module"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/modules/demo"
)

type ModuleMaker func() (module.Module, error)

func (s *Server) ensureModuleMaker() {
	// demo
	s.moduleManager.RegisterMaker(
		demo.ModuleName,
		func() (module.Module, error) {
			return demo.NewModule(s.kc, s.informerFactory)
		},
	)
}
