package errors

import (
	"encoding/json"
	"net/http"
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
