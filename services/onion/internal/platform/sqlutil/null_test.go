package sqlutil

import "testing"

func TestNullString(t *testing.T) {
	if got := NullString(""); got.Valid {
		t.Fatal("empty string should be invalid")
	}
	if got := NullString("value"); !got.Valid || got.String != "value" {
		t.Fatalf("NullString(value) = %#v, want valid value", got)
	}
}

func TestNullInt64(t *testing.T) {
	if got := NullInt64(0); got.Valid {
		t.Fatal("zero ID should be invalid")
	}
	if got := NullInt64(42); !got.Valid || got.Int64 != 42 {
		t.Fatalf("NullInt64(42) = %#v, want valid 42", got)
	}
}
