package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Dialog struct {
	Title   string `json:"title" xml:"title,attr"`
	Message string `json:"message" xml:"message,attr"`
}

type DialogOutput struct {
	Title   string `json:"title" xml:"title,attr"`
	Message string `json:"message" xml:"message,attr"`
}

func (d Dialog) ToOutput() DialogOutput {
	return DialogOutput(d)
}

type ServerError struct {
	Code             int     `json:"code" xml:"code,attr"`
	HTTPCode         int     `json:"-"`
	Message          string  `json:"message" xml:"message,attr"`
	Logoff           *bool   `json:"logoff" xml:"logoff,attr" exhaustruct:"optional"`
	ErrorDialog      *Dialog `json:"dialog" xml:"dialog" exhaustruct:"optional"`
	ProxyReRoute     *string `json:"-" exhaustruct:"optional"`
	RedirectLocation *string `json:"-" exhaustruct:"optional"`
}

func (e ServerError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Code        int     `json:"code" xml:"code,attr"`
		Message     string  `json:"message" xml:"message,attr"`
		Logoff      *bool   `json:"logoff,omitempty" xml:"logoff,attr"`
		ErrorDialog *Dialog `json:"dialog,omitempty" xml:"dialog"`
	}{
		Code:        e.Code,
		Message:     e.Message,
		Logoff:      e.Logoff,
		ErrorDialog: e.ErrorDialog,
	})
}

func (e ServerError) Error() string {
	return e.Message
}

func RedirectServerError(location string) *ServerError {
	return &ServerError{
		RedirectLocation: &location,
	}
}

const ErrorPermission = "You don't have permission to access this resource"

func PermissionDeniedServerError() *ServerError {
	return &ServerError{
		HTTPCode: http.StatusForbidden,
		Code:     http.StatusForbidden,
		Message:  ErrorPermission,
	}
}

type ErrorResponse struct {
	Error  ServerError `json:"error" xml:"error"`
	Status string      `json:"status" xml:"status,attr"`
}

func (e ErrorResponse) GetDialog() (Dialog, bool) {
	if e.Error.ErrorDialog == nil {
		return Dialog{}, false
	}

	return *e.Error.ErrorDialog, true
}

// FromError maps common sentinel errors (e.g. database and standard domain errors)
// to a structured ServerError with correct HTTP status codes.
func FromError(err error, fallbackMessage string) *ServerError {
	if err == nil {
		return nil
	}

	// If it's already a ServerError, return it as-is
	se := &ServerError{}
	if errors.As(err, &se) {
		return se
	}

	errStr := err.Error()

	// Handle sql.ErrNoRows or typical "not found" sentinels
	if errStr == "sql: no rows in result set" || strings.Contains(strings.ToLower(errStr), "not found") {
		return &ServerError{
			HTTPCode: http.StatusNotFound,
			Code:     http.StatusNotFound,
			Message:  errStr,
		}
	}

	// Handle duplicate key / constraint errors
	if strings.Contains(strings.ToLower(errStr), "duplicate") || strings.Contains(strings.ToLower(errStr), "already exists") {
		return &ServerError{
			HTTPCode: http.StatusConflict,
			Code:     http.StatusConflict,
			Message:  errStr,
		}
	}

	// Handle validation errors / bad requests
	if strings.Contains(strings.ToLower(errStr), "invalid") ||
		strings.Contains(strings.ToLower(errStr), "weak") ||
		strings.Contains(strings.ToLower(errStr), "required") ||
		strings.Contains(strings.ToLower(errStr), "validation") {
		return &ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  errStr,
		}
	}

	// Handle permission issues
	if strings.Contains(strings.ToLower(errStr), "denied") || strings.Contains(strings.ToLower(errStr), "unauthorized") {
		return &ServerError{
			HTTPCode: http.StatusForbidden,
			Code:     http.StatusForbidden,
			Message:  errStr,
		}
	}

	// Fallback to internal server error
	msg := fallbackMessage
	if msg == "" {
		msg = errStr
	}

	return &ServerError{
		HTTPCode: http.StatusInternalServerError,
		Code:     http.StatusInternalServerError,
		Message:  msg,
	}
}
