package errors

import (
	"crypto/x509"
	"net/url"

	kubeerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// IsConflict determines if the err is an error which indicates the provided update conflicts.
	IsConflict = kubeerr.IsConflict
	// IsAlreadyExists determines if the err is an error which indicates that a specified resource already exists.
	IsAlreadyExists = kubeerr.IsAlreadyExists
	// IsForbidden determines if err is an error which indicates that the request is forbidden and cannot
	// be completed as requested.
	IsForbidden = kubeerr.IsForbidden
	// NewNotFound returns a new error which indicates that the resource of the kind and the name was not found.
	NewNotFound = kubeerr.NewNotFound
	// IsInvalid determines if the err is an error which indicates the provided resource is not valid.
	IsInvalid = kubeerr.IsInvalid
)

// IsNotFound returns true if the specified error was created by NewNotFound.
func IsNotFound(err error) bool {
	if !kubeerr.IsNotFound(err) {
		return false
	}
	switch err.(type) {
	case kubeerr.APIStatus:
		ke := err.(kubeerr.APIStatus)
		if ke == nil || ke.Status().Details == nil || len(ke.Status().Details.Causes) == 0 {
			return true
		}
		for _, statusCause := range ke.Status().Details.Causes {
			if statusCause.Type == metav1.CauseTypeUnexpectedServerResponse {
				return false
			}
		}
	}
	return true
}

// IsHTTPSUnknownAuthority determines if err is an error which indicates that the certificate issuer is unknown.
func IsHTTPSUnknownAuthority(err error) bool {
	if err == nil {
		return false
	}
	if urlErr := err.(*url.Error); urlErr != nil {
		return isX509UnknownAuthorityError(urlErr.Err)
	}
	return isX509UnknownAuthorityError(err)
}

func isX509UnknownAuthorityError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(x509.UnknownAuthorityError); ok {
		return true
	}
	if _, ok := err.(*x509.UnknownAuthorityError); ok {
		return true
	}
	return false
}
