package models

type ServerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func NewServerError(code int, message string) *ServerError {
	return &ServerError{
		Code:    code,
		Message: message,
	}
}

func (e *ServerError) Error() string {
	return e.Message
}
