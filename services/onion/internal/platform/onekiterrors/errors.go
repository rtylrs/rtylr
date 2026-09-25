package onekiterrors

import (
	"net/http"

	onekithttp "github.com/corezio/onekit/http"
)

const PermissionDeniedMessage = "You don't have permission to access this resource"

type PublicError struct {
	err *onekithttp.Error
}

func (e *PublicError) Error() string {
	return e.err.Error()
}

func (e *PublicError) HTTPStatusCode() int {
	return e.err.HTTPStatusCode()
}

func (e *PublicError) PublicMessage() string {
	return e.err.GetMessage()
}

func (e *PublicError) Unwrap() error {
	return e.err
}

func New(statusCode int, message string) error {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return &PublicError{err: onekithttp.NewError(statusCode, message)}
}

func PermissionDenied() error {
	return New(http.StatusForbidden, PermissionDeniedMessage)
}
