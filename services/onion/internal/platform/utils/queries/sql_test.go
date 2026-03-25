package queries

import (
	"testing"
	"time"
)

func TestString(t *testing.T) {
	res1 := NullStringOrEmpty("test")
	if res1.String != "test" {
		t.Errorf("Expected String value to be %s", "test")
	}
	if res1.Valid != true {
		t.Errorf("Expected Valid value to be %t", true)
	}

	res2 := NullStringOrEmpty("")
	if res2.Valid != false {
		t.Errorf("Expected Valid value to be %t", false)
	}
}

func TestNullString(t *testing.T) {
	res := EmptyNullString()
	if res.Valid != false {
		t.Errorf("Expected Valid value to be %t", false)
	}
}

func TestInt32(t *testing.T) {
	res := Int32(123)
	if res.Int32 != 123 {
		t.Errorf("Expected Int32 value to be %d", 123)
	}
	if res.Valid != true {
		t.Errorf("Expected Valid value to be %t", true)
	}
}

func TestInt64(t *testing.T) {
	res := Int64(123)
	if res.Int64 != 123 {
		t.Errorf("Expected Int64 value to be %d", 123)
	}
	if res.Valid != true {
		t.Errorf("Expected Valid value to be %t", true)
	}
}

func TestTime(t *testing.T) {
	expectedTimeValue := time.Now()
	res1 := NullTimeOrEmpty(expectedTimeValue)
	if res1.Time != expectedTimeValue {
		t.Errorf("Expected Time value to be %s", expectedTimeValue)
	}
	if res1.Valid != true {
		t.Errorf("Expected Valid value to be %t", true)
	}

	res2 := NullTimeOrEmpty(time.Time{})
	if res2.Valid != false {
		t.Errorf("Expected Valid value to be %t", false)
	}
}

func TestNullTime(t *testing.T) {
	res := EmptyNullTime()
	if res.Valid != false {
		t.Errorf("Expected Valid value to be %t", false)
	}
}

func TestBool(t *testing.T) {
	res := Bool(false)
	if res.Bool != false {
		t.Errorf("Expected Bool value to be %t", false)
	}
	if res.Valid != true {
		t.Errorf("Expected Valid value to be %t", true)
	}
}

func TestArgs(t *testing.T) {
	t.Parallel()
	type someType struct {
		A, B, C string
	}

	testCases := []struct {
		name     string
		params   any
		expected []any
	}{
		{
			name:     "not a struct type",
			params:   "test",
			expected: []any{},
		},
		{
			name:     "struct type with values",
			params:   someType{A: "a", B: "b", C: "c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "struct type with default values",
			params:   someType{},
			expected: []any{"", "", ""},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := Args(tt.params)
			if len(res) != len(tt.expected) {
				t.Errorf("Expected length to be %d", len(tt.expected))
			}
			for i, v := range res {
				if v != tt.expected[i] {
					t.Errorf("Expected value to be %v", tt.expected[i])
				}
			}
		})
	}
}
