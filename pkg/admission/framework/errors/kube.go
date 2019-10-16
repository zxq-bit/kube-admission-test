package errors

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	reasonGroupBase        = "admission:"
	ReasonGroupWorkload    = "workload.app." + reasonGroupBase
	ReasonGroupConfig      = "config.app." + reasonGroupBase
	ReasonGroupApplication = "application.app." + reasonGroupBase
	ReasonGroupDemo        = "demo." + reasonGroupBase

	ReasonServerInternal = reasonGroupBase + "InternalServerError"
	ReasonBadRequest     = reasonGroupBase + "BadRequest"
	ReasonRequestTimeout = reasonGroupBase + "RequestTimeout"
	ReasonDemo           = ReasonGroupDemo + "ErrorReasonDemo"
)

type APIStatus = errors.APIStatus
type StatusError = errors.StatusError

func NewKubeErr(code int32, reason string, err error) *StatusError {
	return &errors.StatusError{
		ErrStatus: metav1.Status{
			Code:    code,
			Status:  metav1.StatusFailure,
			Reason:  metav1.StatusReason(reason),
			Message: fmt.Sprintf("%v", err),
		},
	}
}

// NewInternalServerError diff with kube impl by use internal error reason
func NewInternalServerError(err error) *StatusError {
	return NewKubeErr(http.StatusInternalServerError, ReasonServerInternal, err)
}

// NewBadRequest diff with kube impl by use internal error reason
func NewBadRequest(err error) *StatusError {
	return NewKubeErr(http.StatusBadRequest, ReasonBadRequest, err)
}

// NewRequestTimeout diff with kube impl by use internal error reason
func NewRequestTimeout(err error) *StatusError {
	return NewKubeErr(http.StatusRequestTimeout, ReasonRequestTimeout, err)
}
