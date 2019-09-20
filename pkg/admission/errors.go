package admission

import (
	"fmt"
	"net/http"

	orchestrationapi "github.com/caicloud/clientset/pkg/apis/orchestration/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	switch e := err.(type) {
	case errors.APIStatus:
		status := e.Status()
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result:  &status,
		}
	default:
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
}

// NewConflict returns an error indicating the item can't be updated as provided.
func NewConflict(format string, a ...interface{}) error {
	return &errors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusConflict,
			Reason: metav1.StatusReasonConflict,
			Details: &metav1.StatusDetails{
				Group: orchestrationapi.GroupName,
				Kind:  "Application",
			},
			Message: "Conflict: " + fmt.Sprintf(format, a...),
		},
	}
}

// NewBadRequest creates an error that indicates that the request is invalid and can not be processed.
func NewBadRequest(format string, a ...interface{}) error {
	return &errors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusBadRequest,
			Reason: metav1.StatusReasonBadRequest,
			Details: &metav1.StatusDetails{
				Group: orchestrationapi.GroupName,
				Kind:  "Application",
			},
			Message: "BadRequest: " + fmt.Sprintf(format, a...),
		},
	}
}

// NewInternalError returns an error indicating the item is invalid and cannot be processed.
func NewInternalError(format string, a ...interface{}) error {
	return &errors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusInternalServerError,
			Reason: metav1.StatusReasonInternalError,
			Details: &metav1.StatusDetails{
				Group: orchestrationapi.GroupName,
				Kind:  "Application",
			},
			Message: fmt.Sprintf(format, a...),
		},
	}
}
