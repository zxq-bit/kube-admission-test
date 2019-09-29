package errors

import (
	"fmt"
)

var (
	ErrContextEnded = fmt.Errorf("context ended")
)
