package http

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/proto"

	"github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	validateInstance          = validator.New()
	protoValidator            = protovalidate.GlobalValidator
	errMultipleJSONBodyValues = stdErrors.New("request body must contain a single JSON value")
)

type MiddlewareFunc func(next http.Handler) http.Handler

func ApplyMiddlewares(r *chi.Mux, middlewares []MiddlewareFunc) {
	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}

type HandlerFunc[Req any, Res any] interface {
	Serve(ctx context.Context, req *Req) (Res, *errors.ServerError)
}

type ValidatorFunc[Req any] interface {
	Validate(ctx context.Context, req *Req) *errors.ServerError
}

func RegisterGET[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Get(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterGETWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Get(path, wrapHandler(ctx, handler, true))
	}
}

// RegisterGETWithValidationAndMiddleware is the GET counterpart to
// RegisterPOSTWithValidationAndMiddleware: a validated GET endpoint with
// per-route middleware (e.g. a tighter rate limit on an anonymous lookup like
// /get-email-exists). Middlewares run in order around the handler — index 0 is
// outermost.
func RegisterGETWithValidationAndMiddleware[Req, Res any](
	ctx context.Context,
	router *chi.Mux,
	paths []string,
	handler HandlerFunc[Req, Res],
	middlewares ...MiddlewareFunc,
) {
	for _, path := range paths {
		var final http.Handler = wrapHandler(ctx, handler, true)
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		router.Method(http.MethodGet, path, final)
	}
}

func RegisterPOST[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Post(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterPOSTWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Post(path, wrapHandler(ctx, handler, true))
	}
}

// RegisterPOSTWithValidationAndMiddleware is the per-route
// middleware-aware variant. Use it for endpoints that need an extra
// layer (the main case today is per-route rate limits on
// /auth/login, /auth/register, /recruit/public/apply). Middlewares
// run in order around the handler — index 0 is outermost.
func RegisterPOSTWithValidationAndMiddleware[Req, Res any](
	ctx context.Context,
	router *chi.Mux,
	paths []string,
	handler HandlerFunc[Req, Res],
	middlewares ...MiddlewareFunc,
) {
	for _, path := range paths {
		var final http.Handler = wrapHandler(ctx, handler, true)
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		router.Method(http.MethodPost, path, final)
	}
}

func RegisterPATCH[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Patch(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterPATCHWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Patch(path, wrapHandler(ctx, handler, true))
	}
}

func RegisterPUT[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Put(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterPUTWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Put(path, wrapHandler(ctx, handler, true))
	}
}

func RegisterDELETE[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Delete(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterDELETEWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Delete(path, wrapHandler(ctx, handler, true))
	}
}

func wrapHandler[Req, Res any](ctx context.Context, handler HandlerFunc[Req, Res], validate bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		switch r.Method {
		case "POST", "PUT", "PATCH":
			if !decodeJSONRequestBody(ctx, w, r, &req) {
				return
			}
		case "DELETE":
			// DELETE can ride with either a JSON body (the FE pattern
			// for {"id": N}) or query params. RFC 9110 doesn't forbid
			// a body, just discourages relying on it. Prefer body when
			// present so handlers don't have to special-case the FE
			// shape; fall back to query params otherwise.
			if r.Body != nil && r.ContentLength != 0 {
				if !decodeJSONRequestBody(ctx, w, r, &req) {
					return
				}
			} else {
				parseQueryParams(r, &req)
			}
		case "GET":
			parseQueryParams(r, &req)
		}

		reqCtx := r.Context()

		if validate {
			if vErr := validateRequest(reqCtx, handler, &req); vErr != nil {
				writeServerError(reqCtx, w, vErr)

				return
			}
		}

		res, serverError := handler.Serve(reqCtx, &req)
		if serverError != nil {
			writeServerError(reqCtx, w, serverError)

			return
		}

		if err := validateProtoMessages(res); err != nil {
			logger.Ctx(reqCtx).Error("Response proto validation failed", logger.Error(err))
			writeServerError(reqCtx, w, &errors.ServerError{
				HTTPCode: http.StatusInternalServerError,
				Code:     http.StatusInternalServerError,
				Message:  "Response validation failed",
			})

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Ctx(ctx).Error("Failed to encode response", logger.Error(err))
		}
	}
}

func decodeJSONRequestBody(ctx context.Context, w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := DecodeJSONRequestBody(w, r, dst); err != nil {
		writeBadRequest(ctx, w)

		return false
	}

	return true
}

// DecodeJSONRequestBody reads at most one JSON value from the request body.
// Empty bodies are accepted so legacy no-payload endpoints keep working.
func DecodeJSONRequestBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20) // 4 MiB

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errMultipleJSONBodyValues
		}
		return err
	}

	return nil
}

func writeBadRequest(ctx context.Context, w http.ResponseWriter) {
	logger.Ctx(ctx).Warn("Bad request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := errors.ErrorResponse{
		Error: errors.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid request",
		},
		Status: "error",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(err))
	}
}

func validateRequest[Req any](ctx context.Context, handler interface{}, req *Req) *errors.ServerError {
	// First run structural validation using struct tags if present
	if err := validateInstance.Struct(req); err != nil {
		var msg string
		var validationErrs validator.ValidationErrors
		if stdErrors.As(err, &validationErrs) {
			var fields []string
			for _, ve := range validationErrs {
				fields = append(fields, fmt.Sprintf("%s (%s)", ve.Field(), ve.Tag()))
			}
			msg = "Validation failed for fields: " + strings.Join(fields, ", ")
		} else {
			msg = "Validation error: " + err.Error()
		}

		return &errors.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  msg,
		}
	}

	if err := validateProtoMessages(req); err != nil {
		return &errors.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Proto validation failed: " + err.Error(),
		}
	}

	// Then run custom logic validation if the handler implements ValidatorFunc
	if v, ok := handler.(ValidatorFunc[Req]); ok {
		if err := v.Validate(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func validateProtoMessages(value any) error {
	seen := make(map[uintptr]struct{})
	return walkProtoMessages(reflect.ValueOf(value), seen)
}

func walkProtoMessages(value reflect.Value, seen map[uintptr]struct{}) error {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		return walkProtoMessages(value.Elem(), seen)
	}

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		ptr := value.Pointer()
		if _, ok := seen[ptr]; ok {
			return nil
		}
		seen[ptr] = struct{}{}

		if value.CanInterface() {
			if msg, ok := value.Interface().(proto.Message); ok {
				return protoValidator.Validate(msg)
			}
		}

		return walkProtoMessages(value.Elem(), seen)
	}

	if value.CanAddr() {
		ptr := value.Addr()
		if ptr.CanInterface() {
			if msg, ok := ptr.Interface().(proto.Message); ok {
				return protoValidator.Validate(msg)
			}
		}
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			if !field.CanInterface() && !field.CanAddr() {
				continue
			}
			if err := walkProtoMessages(field, seen); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := walkProtoMessages(value.Index(i), seen); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if err := walkProtoMessages(value.MapIndex(key), seen); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseQueryParams[Req any](r *http.Request, req *Req) {
	query := r.URL.Query()
	if len(query) == 0 {
		return
	}

	queryMap := make(map[string]any)

	for key, values := range query {
		val := values[0]
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			queryMap[key] = n
		} else if f, err := strconv.ParseFloat(val, 64); err == nil {
			queryMap[key] = f
		} else if val == "true" || val == "false" {
			queryMap[key] = val == "true"
		} else {
			queryMap[key] = val
		}
	}

	queryJSON, err := json.Marshal(queryMap)
	if err != nil {
		return
	}

	_ = json.Unmarshal(queryJSON, req) //nolint:errcheck // best-effort query param parsing
}

func writeServerError(ctx context.Context, w http.ResponseWriter, serverError *errors.ServerError) {
	w.Header().Set("Content-Type", "application/json")

	// Most call sites set Code (the app-level status) and leave HTTPCode
	// blank. Falling back to Code keeps a 400 from being silently
	// promoted to a 500 just because HTTPCode wasn't filled in. Only
	// promote to 500 when neither is set.
	if serverError.HTTPCode == 0 {
		if serverError.Code >= 400 && serverError.Code < 600 {
			serverError.HTTPCode = serverError.Code
		} else {
			serverError.HTTPCode = http.StatusInternalServerError
		}
	}

	w.WriteHeader(serverError.HTTPCode)

	response := errors.ErrorResponse{
		Error:  *serverError,
		Status: "error",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(err))
	}
}
