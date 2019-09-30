package errors

import (
	"fmt"
)

var (
	ErrContextEnded          = fmt.Errorf("context ended")
	ErrProcessorIsNil        = fmt.Errorf("processor is nil")
	ErrEmptyGVR              = fmt.Errorf("groupVersionResource is empty")
	ErrBadTimeoutSecond      = fmt.Errorf("timeout second must between 0 and 30 seconds")
	ErrNilRawExtensionParser = fmt.Errorf("RawExtensionParser is nil")
	ErrNilReviewer           = fmt.Errorf("reviewer is nil")
	ErrNilRuntimeObject      = fmt.Errorf("runtime.Object is nil")
	ErrRuntimeObjectBadType  = fmt.Errorf("runtime.Object is not required type")
)
