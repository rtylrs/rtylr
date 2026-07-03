package http

import (
	"errors"
	"net/http"
	"strings"

	onekithttp "github.com/1homsi/onekit/http"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"google.golang.org/protobuf/proto"
)

// OneKitErrorHandler adapts generated oneKit HTTP servers to Onion's existing
// JSON error envelope (writeServerError), so oneKit-backed and hand-written
// routes return indistinguishable error responses to clients.
func OneKitErrorHandler(w http.ResponseWriter, r *http.Request, err error) proto.Message {
	writeServerError(r.Context(), w, serverErrorFromOneKit(err))
	return nil
}

func serverErrorFromOneKit(err error) *errorModels.ServerError {
	if serverErr, ok := errors.AsType[*errorModels.ServerError](err); ok {
		return serverErr
	}

	if validationErr, ok := errors.AsType[*onekithttp.ValidationError](err); ok {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  oneKitValidationMessage(validationErr),
		}
	}

	if oneKitErr, ok := errors.AsType[*onekithttp.Error](err); ok && oneKitErr.GetMessage() != "" {
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

func oneKitValidationMessage(err *onekithttp.ValidationError) string {
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
