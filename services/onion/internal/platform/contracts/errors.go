package contracts

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	onekithttp "github.com/1homsi/onekit/http"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ServiceError carries Onion's HTTP error envelope through generated oneKit handlers.
type ServiceError struct {
	serverErr *errorModels.ServerError
}

func NewServiceError(err *errorModels.ServerError) error {
	if err == nil {
		return nil
	}
	return &ServiceError{serverErr: err}
}

func NewStatusError(status int, message string) error {
	return NewServiceError(&errorModels.ServerError{HTTPCode: status, Code: status, Message: message})
}

func (e *ServiceError) Error() string {
	if e == nil || e.serverErr == nil {
		return "request failed"
	}
	return e.serverErr.Message
}

func (e *ServiceError) ProtoReflect() protoreflect.Message {
	return (&onekithttp.Error{Message: e.Error()}).ProtoReflect()
}

func ErrorHandler(w http.ResponseWriter, _ *http.Request, err error) proto.Message {
	serverErr := ServerErrorFrom(err)
	status := serverErr.HTTPCode
	if status == 0 {
		status = serverErr.Code
	}
	if status == 0 {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorModels.ErrorResponse{Error: *serverErr, Status: "error"})
	return nil
}

func ServerErrorFrom(err error) *errorModels.ServerError {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) && serviceErr.serverErr != nil {
		return serviceErr.serverErr
	}

	var validationErr *onekithttp.ValidationError
	if errors.As(err, &validationErr) {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  validationErrorMessage(validationErr),
		}
	}

	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) && oneKitErr.GetMessage() != "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  oneKitErr.GetMessage(),
		}
	}

	return &errorModels.ServerError{
		HTTPCode: http.StatusInternalServerError,
		Code:     http.StatusInternalServerError,
		Message:  "Request failed",
	}
}

func validationErrorMessage(err *onekithttp.ValidationError) string {
	violations := err.GetViolations()
	if len(violations) == 0 {
		return "Invalid request"
	}

	parts := make([]string, 0, len(violations))
	for _, violation := range violations {
		field := violation.GetField()
		description := violation.GetDescription()
		if field == "" {
			parts = append(parts, description)
			continue
		}
		if description == "" {
			parts = append(parts, field)
			continue
		}
		parts = append(parts, field+": "+description)
	}

	return strings.Join(parts, "; ")
}
