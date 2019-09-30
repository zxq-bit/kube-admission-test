package errors

import (
	"fmt"
)

var (
	ErrContextEnded          = fmt.Errorf("context ended")
	ErrProcessorIsNil        = fmt.Errorf("processor is nil")
	ErrEmptyGVR              = fmt.Errorf("empty GroupVersionResource")
	ErrNoHandlerMakerGVR     = fmt.Errorf("handler maker for this GroupVersionResource not registered")
	ErrBadOperationType      = fmt.Errorf("bad operation type, not in list")
	ErrBadTimeoutSecond      = fmt.Errorf("timeout second should be 0 or positive integer")
	ErrNilRawExtensionParser = fmt.Errorf("RawExtensionParser is nil")
	ErrNilReviewer           = fmt.Errorf("reviewer is nil")
	ErrNilRuntimeObject      = fmt.Errorf("runtime.Object is nil")
	ErrRuntimeObjectBadType  = fmt.Errorf("runtime.Object is not required type")
)
