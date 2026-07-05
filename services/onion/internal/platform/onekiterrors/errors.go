package onekiterrors

import (
	"net/http"

	onekithttp "github.com/corezio/onekit/http"
)

const PermissionDeniedMessage = "You don't have permission to access this resource"

func New(statusCode int, message string) error {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return onekithttp.NewError(statusCode, message)
}

func PermissionDenied() error {
	return New(http.StatusForbidden, PermissionDeniedMessage)
}
