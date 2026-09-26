package onekiterrors

import "net/http"

const PermissionDeniedMessage = "You don't have permission to access this resource"

// Error is the HTTP error type handlers return. The generated onekit server
// reads HTTPStatusCode and PublicMessage through interfaces, so this package
// carries no runtime dependency on onekit itself.
type Error struct {
	statusCode int
	message    string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) HTTPStatusCode() int {
	return e.statusCode
}

func (e *Error) GetMessage() string {
	return e.message
}

func (e *Error) GetStatusCode() int32 {
	return int32(e.statusCode)
}

// PublicMessage is the message safe to send to the client; the generated
// server prefers it over Error() when present.
func (e *Error) PublicMessage() string {
	return e.message
}

func NewError(statusCode int, message string) *Error {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return &Error{statusCode: statusCode, message: message}
}

// PublicError keeps the historical wrapper so callers can Unwrap to the
// concrete *Error when asserting with errors.As.
type PublicError struct {
	err *Error
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
	return &PublicError{err: NewError(statusCode, message)}
}

func PermissionDenied() error {
	return New(http.StatusForbidden, PermissionDeniedMessage)
}
