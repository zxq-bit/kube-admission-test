package server

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/models/app"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/models/demo"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
)

func (s *Server) initModelsAndProcessors() error {
	// model
	appModel, e := app.NewModel(s.kc, s.informerFactory)
	if e != nil {
		return e
	}
	e = s.modelCollection.Register(appModel)
	if e != nil {
		return e
	}
	// processor
	// pod
	// pod create
	s.configCollection.PodConfig.SetTimeout(arv1b1.Create, 0)
	s.configCollection.PodConfig.Register(arv1b1.Create,
		appModel.GetPodProcessor(app.ProcessorNamePodGPUVisible),
	)

	// cm
	// cm create
	s.configCollection.ConfigMapConfig.SetTimeout(arv1b1.Create, 0)
	s.configCollection.ConfigMapConfig.Register(arv1b1.Create,
		demo.GetConfigMapStaticProcessor(),
	)
	s.configCollection.ConfigMapConfig.SetTimeout(arv1b1.Update, 0)
	s.configCollection.ConfigMapConfig.Register(arv1b1.Update,
		demo.GetConfigMapStaticProcessor(),
	)
	// model
	return nil
}
