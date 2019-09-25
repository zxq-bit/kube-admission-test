package server

import (
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
)

func (s *Server) initModelsAndProcessors() error {
	// model
	e := s.modelCollection.Register()
	if e != nil {
		return e
	}
	// processor
	s.processorConfigCollection.PodConfig.SetTimeout(arv1b1.Create, 0)
	s.processorConfigCollection.PodConfig.Register(arv1b1.Create)
	return nil
}
