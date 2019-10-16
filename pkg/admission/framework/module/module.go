package module

type Module interface {
	// Name get name of this module
	Name() string
	// GetProcessor get module processor by name, return nil if not exist
	GetProcessor(name string) interface{}
	// Start do custom start
	Start(stopCh <-chan struct{})
}
