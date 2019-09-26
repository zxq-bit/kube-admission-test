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
	s.configCollection.PodConfig.SetTimeout(arv1b1.Create, 0)
	s.configCollection.PodConfig.Register(arv1b1.Create)
	return nil
}
