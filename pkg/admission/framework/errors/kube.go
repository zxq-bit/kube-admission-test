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
	ReasonDemo           = ReasonGroupDemo + "ErrorReasonDemo"
)

func NewKubeErr(code int32, reason string, err error) *errors.StatusError {
	return &errors.StatusError{
		ErrStatus: metav1.Status{
			Code:    code,
			Status:  metav1.StatusFailure,
			Reason:  metav1.StatusReason(reason),
			Message: fmt.Sprintf("%v", err),
		},
	}
}

func NewInternalServerError(err error) *errors.StatusError {
	return NewKubeErr(http.StatusInternalServerError, ReasonServerInternal, err)
}

func NewBadRequest(err error) *errors.StatusError {
	return NewKubeErr(http.StatusBadRequest, ReasonBadRequest, err)
}
