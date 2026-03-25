package converter

import (
	"context"
	"reflect"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

func Convert[T any](domain interface{}) T {
	var result T

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr {
		elemType := resultType.Elem()
		elemValue := reflect.New(elemType)
		convertStruct(domain, elemValue.Interface())

		convertedResult, ok := elemValue.Interface().(T)
		if !ok {
			logger.Ctx(context.Background()).Error("Type assertion failed",
				logger.String("expected_type", reflect.TypeOf(result).String()),
				logger.String("actual_type", elemValue.Type().String()),
			)
		}

		return convertedResult
	}

	convertStruct(domain, &result)

	return result
}

func ConvertSlice[T any](domainSlice interface{}) []T {
	domainValue := reflect.ValueOf(domainSlice)
	if domainValue.Kind() != reflect.Slice {
		return nil
	}

	resultSlice := make([]T, domainValue.Len())
	for i := 0; i < domainValue.Len(); i++ {
		item := domainValue.Index(i).Interface()
		resultSlice[i] = Convert[T](item)
	}

	return resultSlice
}

func convertStruct(src, dst interface{}) {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	if dstValue.Kind() != reflect.Ptr || dstValue.Elem().Kind() != reflect.Struct {
		return
	}

	dstElem := dstValue.Elem()
	dstType := dstElem.Type()

	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstType.Field(i)
		dstFieldValue := dstElem.Field(i)

		if !dstFieldValue.CanSet() {
			continue
		}

		if srcField := findFieldByCaseInsensitiveName(srcValue, dstField.Name); srcField.IsValid() {
			setFieldValue(srcField, dstFieldValue)
		} else {
			logger.Ctx(context.Background()).Error("Field not found in source",
				logger.String("field_name", dstField.Name),
				logger.String("dst_type", dstType.Name()),
				logger.String("src_type", srcValue.Type().Name()),
			)
		}
	}
}

func setFieldValue(src, dst reflect.Value) {
	if !dst.CanSet() {
		return
	}

	srcType := src.Type()
	dstType := dst.Type()

	if srcType == dstType {
		dst.Set(src)

		return
	}

	if srcType.ConvertibleTo(dstType) {
		dst.Set(src.Convert(dstType))

		return
	}

	logger.Ctx(context.Background()).Error("Field type conversion failed",
		logger.String("src_type", srcType.String()),
		logger.String("dst_type", dstType.String()),
	)
}

func findFieldByCaseInsensitiveName(srcValue reflect.Value, fieldName string) reflect.Value {
	srcType := srcValue.Type()

	if srcType.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return reflect.Value{}
		}

		srcValue = srcValue.Elem()
		srcType = srcValue.Type()
	}

	if srcType.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	for i := 0; i < srcType.NumField(); i++ {
		srcFieldName := srcType.Field(i).Name
		if strings.EqualFold(srcFieldName, fieldName) {
			return srcValue.Field(i)
		}
	}

	return reflect.Value{}
}
