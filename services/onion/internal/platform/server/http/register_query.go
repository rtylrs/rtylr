package http

import (
	"bytes"
	"context"
	"encoding"
	stdErrors "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	onekithttp "github.com/corezio/onekit/http"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/voxire/onion/internal/platform/logger"
)

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
