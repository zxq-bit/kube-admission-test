package model

type Model interface {
	// Name get name of this model
	Name() string
	// GetProcessor get model processor by name, return nil if not exist
	GetProcessor(name string) interface{}
	// Start do custom start
	Start(stopCh <-chan struct{})
}
