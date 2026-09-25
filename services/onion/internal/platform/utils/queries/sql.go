package queries

import (
	"database/sql"
	"reflect"
	"time"
)

// NullStringOrEmpty craft sql.NullString object based on the input, if value is empty string set to null.
func NullStringOrEmpty(value string) sql.NullString {
	if value == "" {
		return EmptyNullString()
	}

	return sql.NullString{
		String: value,
		Valid:  true,
	}
}

// EmptyNullString String craft empty sql.NullString.
func EmptyNullString() sql.NullString {
	return sql.NullString{
		Valid: false,
	}
}

// Int32 craft sql.NullInt32 object based on the input.
func Int32(value int32) sql.NullInt32 {
	return sql.NullInt32{
		Int32: value,
		Valid: true,
	}
}

// Int64 craft sql.NullInt64 object based on the input.
func Int64(value int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: value,
		Valid: true,
	}
}

// NullTimeOrEmpty craft sql.NullTime object based on the input, if time is zero set to null.
func NullTimeOrEmpty(value time.Time) sql.NullTime {
	if value.IsZero() {
		return EmptyNullTime()
	}

	return sql.NullTime{
		Time:  value,
		Valid: true,
	}
}

// EmptyNullTime String craft empty sql.NullTime.
func EmptyNullTime() sql.NullTime {
	return sql.NullTime{
		Valid: false,
	}
}
func Bool(value bool) sql.NullBool {
	return sql.NullBool{
		Bool:  value,
		Valid: true,
	}
}

// Args takes a struct as a parameter and returns a slice of 'any'.
// It uses reflection to iterate over the fields of the given struct
// and adds each field's value to the slice in the order they are declared in the struct.
//
// # This function is useful when you need to transform a struct representation of a query parameter into a slice of any
//
// This function only works with structs. Passing a non-struct parameter will simply return an empty slice.
//
// Params:
//   - params: A struct from which the fields will be extracted.
//
// Returns:
//
//	A slice of 'any', containing the values of the struct's fields in order.
//
// Example:
//
//	type MyStruct struct {
//	    Name string
//	    Age  int
//	}
//
//	func main() {
//	    s := MyStruct{Name: "Alice", Age: 30}
//	    fields := Args(s)
//	    for _, field := range fields {
//	        fmt.Println(field)
//	    }
//	    // Output:
//	    // Alice
//	    // 30
//	}
func Args(params any) []any {
	v := reflect.ValueOf(params)

	if v.Kind() != reflect.Struct {
		return []any{}
	}

	args := make([]any, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		args[i] = v.Field(i).Interface()
	}

	return args
}
