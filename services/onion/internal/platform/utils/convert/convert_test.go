package convert

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestToString(t *testing.T) {
	if got := ToString(sql.NullString{}); got != "" {
		t.Errorf("invalid -> %q", got)
	}
	if got := ToString(sql.NullString{String: "hi", Valid: true}); got != "hi" {
		t.Errorf("valid -> %q", got)
	}
}

func TestToBool(t *testing.T) {
	if ToBool(sql.NullBool{}) != false {
		t.Error("invalid -> should be false")
	}
	if !ToBool(sql.NullBool{Bool: true, Valid: true}) {
		t.Error("valid true -> should be true")
	}
}

func TestToInt32_Int64_Float64(t *testing.T) {
	if ToInt32(sql.NullInt32{}) != 0 {
		t.Error("invalid int32 should be 0")
	}
	if ToInt32(sql.NullInt32{Int32: 42, Valid: true}) != 42 {
		t.Error("valid int32 should pass through")
	}
	if ToInt64(sql.NullInt64{Int64: 9001, Valid: true}) != 9001 {
		t.Error("valid int64 should pass through")
	}
	if ToFloat64(sql.NullFloat64{Float64: 3.14, Valid: true}) != 3.14 {
		t.Error("valid float64 should pass through")
	}
}

func TestToStringSlice(t *testing.T) {
	if got := ToStringSlice(nil); got != nil {
		t.Errorf("nil -> %v", got)
	}
	if got := ToStringSlice(json.RawMessage(`["a","b","c"]`)); len(got) != 3 || got[1] != "b" {
		t.Errorf("got %v", got)
	}
	if got := ToStringSlice(json.RawMessage(`{"not":"a list"}`)); got != nil {
		t.Errorf("malformed -> %v", got)
	}
}

func TestBoolToString(t *testing.T) {
	if BoolToString(true) != "1" || BoolToString(false) != "0" {
		t.Error("BoolToString")
	}
}

func TestIntToString(t *testing.T) {
	if Int32ToString(7) != "7" || IntToString(42) != "42" {
		t.Error("IntToString")
	}
}
