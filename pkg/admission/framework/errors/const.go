package errors

import (
	"fmt"
)

var (
	ErrContextEnded        = fmt.Errorf("context ended")
	ErrProcessorIsNil      = fmt.Errorf("processor is nil")
	ErrEmptyGVR            = fmt.Errorf("groupVersionResource is empty")
	ErrBadTimeoutSecond    = fmt.Errorf("timeout second must between 0 and 30 seconds")
	ErrWrongRuntimeObjects = fmt.Errorf("both old and new runtime.Object is nil or not required type")
)
