package http

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"buf.build/go/protovalidate"
	onekithttp "github.com/corezio/onekit/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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
	Serve(ctx context.Context, req *Req) (Res, error)
}

type ValidatorFunc[Req any] interface {
	Validate(ctx context.Context, req *Req) error
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
	validateResponses := responseProtoValidationEnabled()

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
				if err := parseQueryParams(r, &req); err != nil {
					WriteOneKitError(r.Context(), w, onekiterrors.New(http.StatusBadRequest, err.Error()))
					return
				}
			}
		case "GET":
			if err := parseQueryParams(r, &req); err != nil {
				WriteOneKitError(r.Context(), w, onekiterrors.New(http.StatusBadRequest, err.Error()))
				return
			}
		}

		reqCtx := r.Context()

		if validate {
			if vErr := validateRequest(reqCtx, handler, &req); vErr != nil {
				WriteOneKitError(reqCtx, w, vErr)

				return
			}
		}

		res, serverError := handler.Serve(reqCtx, &req)
		if serverError != nil {
			WriteOneKitError(reqCtx, w, serverError)

			return
		}

		if validateResponses {
			if err := validateProtoMessages(res); err != nil {
				logger.Ctx(reqCtx).Error("Response proto validation failed", logger.Error(err))
				WriteOneKitError(reqCtx, w, onekiterrors.New(http.StatusInternalServerError, "Response validation failed"))

				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Ctx(ctx).Error("Failed to encode response", logger.Error(err))
		}
	}
}

func responseProtoValidationEnabled() bool {
	return envFlag("VALIDATE_RESPONSE_PROTOS") || envFlag("RESPONSE_PROTO_VALIDATION")
}

func envFlag(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func decodeJSONRequestBody(ctx context.Context, w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := DecodeJSONRequestBody(w, r, dst); err != nil {
		if isRequestBodyTooLarge(err) {
			writeRequestEntityTooLarge(ctx, w)
		} else {
			writeBadRequest(ctx, w)
		}

		return false
	}

	return true
}

// DecodeJSONRequestBody reads at most one JSON value from the request body.
// Empty bodies are accepted so legacy no-payload endpoints keep working.
func DecodeJSONRequestBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if !requestBodyLimitExempt(r.URL.Path) {
		r.Body = http.MaxBytesReader(w, r.Body, DefaultMaxRequestBodyBytes)
	}

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

	WriteOneKitError(ctx, w, onekiterrors.New(http.StatusBadRequest, "Invalid request"))
}

func validateRequest[Req any](ctx context.Context, handler interface{}, req *Req) error {
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

		return onekiterrors.New(http.StatusBadRequest, msg)
	}

	if err := validateProtoMessages(req); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "Proto validation failed: "+err.Error())
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
			// onek-generated messages implement Validate() error directly
			// (see gengo.GenerateValidation) instead of being a proto.Message.
			// Unlike protovalidate, it only checks the message's own fields,
			// not nested messages - so a passing Validate() still falls
			// through to walk this value's fields below, to reach any nested
			// message's own Validate(). proto.Message still applies to any
			// service still generated by the old protoc plugin.
			if v, ok := value.Interface().(interface{ Validate() error }); ok {
				if err := v.Validate(); err != nil {
					return err
				}
			} else if msg, ok := value.Interface().(proto.Message); ok {
				return protoValidator.Validate(msg)
			}
		}

		return walkProtoMessages(value.Elem(), seen)
	}

	if value.CanAddr() {
		ptr := value.Addr()
		if ptr.CanInterface() {
			if v, ok := ptr.Interface().(interface{ Validate() error }); ok {
				if err := v.Validate(); err != nil {
					return err
				}
			} else if msg, ok := ptr.Interface().(proto.Message); ok {
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

func parseQueryParams[Req any](r *http.Request, req *Req) error {
	if r == nil || r.URL == nil || req == nil {
		return nil
	}

	query := r.URL.Query()
	if len(query) == 0 {
		return nil
	}

	target := reflect.ValueOf(req)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return nil
	}
	return assignQueryParams(target.Elem(), query)
}

func assignQueryParams(value reflect.Value, query url.Values) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			if !value.CanSet() {
				return nil
			}
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := valueType.Field(i)
		fieldValue := value.Field(i)
		if structField.PkgPath != "" && !structField.Anonymous {
			continue
		}

		fieldName, ok := queryParamFieldName(structField)
		if !ok {
			continue
		}

		if structField.Anonymous && fieldName == "" {
			if err := assignQueryParams(fieldValue, query); err != nil {
				return err
			}
			continue
		}

		values, exists := queryValuesForField(query, fieldName)
		if !exists || len(values) == 0 {
			continue
		}
		if err := setQueryField(fieldValue, values, fieldName); err != nil {
			return err
		}
	}

	return nil
}

func queryParamFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}

	if tag != "" {
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return name, true
		}
	}

	if field.Anonymous {
		return "", true
	}

	return field.Name, true
}

func queryValuesForField(query url.Values, fieldName string) ([]string, bool) {
	if values, exists := query[fieldName]; exists {
		return values, true
	}
	if fieldName == "" {
		return nil, false
	}

	lowerFirst := strings.ToLower(fieldName[:1]) + fieldName[1:]
	if values, exists := query[lowerFirst]; exists {
		return values, true
	}

	lower := strings.ToLower(fieldName)
	if values, exists := query[lower]; exists {
		return values, true
	}

	return nil, false
}

func setQueryField(field reflect.Value, values []string, fieldName string) error {
	if !field.CanSet() || len(values) == 0 {
		return nil
	}

	if field.Kind() == reflect.Pointer {
		elem := reflect.New(field.Type().Elem())
		if err := setQueryField(elem.Elem(), values, fieldName); err != nil {
			return err
		}
		field.Set(elem)
		return nil
	}

	if field.Kind() == reflect.Slice {
		slice := reflect.MakeSlice(field.Type(), 0, len(values))
		for _, value := range values {
			elem := reflect.New(field.Type().Elem()).Elem()
			if err := setQueryField(elem, []string{value}, fieldName); err != nil {
				return err
			}
			slice = reflect.Append(slice, elem)
		}
		field.Set(slice)
		return nil
	}

	return setQueryScalar(field, values[0], fieldName)
}

func setQueryScalar(field reflect.Value, value, fieldName string) error {
	if !field.CanSet() {
		return nil
	}

	if field.CanAddr() && field.Addr().CanInterface() {
		if unmarshaler, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := unmarshaler.UnmarshalText([]byte(value)); err != nil {
				logInvalidQueryValue(fieldName, value, err)
				return fmt.Errorf("invalid query parameter %q: %w", fieldName, err)
			}
			return nil
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			logInvalidQueryValue(fieldName, value, err)
			return fmt.Errorf("invalid query parameter %q: %w", fieldName, err)
		}
		field.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			logInvalidQueryValue(fieldName, value, err)
			return fmt.Errorf("invalid query parameter %q: %w", fieldName, err)
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			logInvalidQueryValue(fieldName, value, err)
			return fmt.Errorf("invalid query parameter %q: %w", fieldName, err)
		}
		field.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			logInvalidQueryValue(fieldName, value, err)
			return fmt.Errorf("invalid query parameter %q: %w", fieldName, err)
		}
		field.SetFloat(parsed)
	case reflect.Interface:
		field.Set(reflect.ValueOf(parseQueryPrimitive(value)))
	default:
		return fmt.Errorf("unsupported query parameter %q", fieldName)
	}

	return nil
}

func logInvalidQueryValue(fieldName, value string, err error) {
	logger.Global().Warn("invalid query parameter value",
		logger.String("parameter", fieldName),
		logger.String("value", value),
		logger.Error(err),
	)
}

func parseQueryPrimitive(value string) any {
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parsed
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed
	}
	if value == "true" || value == "false" {
		return value == "true"
	}
	return value
}

func WriteOneKitError(ctx context.Context, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	var response proto.Message

	var validationErr *onekithttp.ValidationError
	if stdErrors.As(err, &validationErr) {
		statusCode = http.StatusBadRequest
		response = validationErr
	} else {
		var oneKitErr *onekithttp.Error
		if !stdErrors.As(err, &oneKitErr) {
			logger.Ctx(ctx).Error("Unhandled internal error", logger.Error(err))
			oneKitErr = onekithttp.NewError(http.StatusInternalServerError, "Internal server error")
		}
		statusCode = oneKitErr.HTTPStatusCode()
		response = oneKitErr
	}

	payload, marshalErr := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(response)
	if marshalErr != nil {
		logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(marshalErr))
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(statusCode)
	if _, writeErr := io.Copy(w, bytes.NewReader(payload)); writeErr != nil {
		logger.Ctx(ctx).Error("Failed to write error response", logger.Error(writeErr))
	}
}
