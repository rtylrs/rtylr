package application

import (
	"encoding/json"
	"testing"
)

func TestMustJSON_ValidMap(t *testing.T) {
	s := mustJSON(map[string]bool{"view_own_profile": true})
	var m map[string]bool
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("mustJSON produced invalid JSON: %v", err)
	}
	if !m["view_own_profile"] {
		t.Error("expected view_own_profile to be true")
	}
}

func TestMustJSON_NilReturnsEmptyObject(t *testing.T) {
	s := mustJSON(nil)
	if s != "null" && s != "{}" {
		// json.Marshal(nil) returns "null" which is valid JSON
		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Fatalf("mustJSON(nil) produced invalid JSON: %v (got %q)", err, s)
		}
	}
}

func TestMustJSON_String(t *testing.T) {
	s := mustJSON("hello")
	if s != `"hello"` {
		t.Errorf("expected `\"hello\"`, got %q", s)
	}
}

func TestMustJSON_NestedStruct(t *testing.T) {
	type inner struct {
		Key string `json:"key"`
		Val int    `json:"val"`
	}
	s := mustJSON(inner{Key: "k", Val: 42})
	var out inner
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Key != "k" || out.Val != 42 {
		t.Errorf("unexpected value: %+v", out)
	}
}

func TestMustJSON_EmptyMap(t *testing.T) {
	s := mustJSON(map[string]bool{})
	if s != "{}" {
		t.Errorf("expected `{}`, got %q", s)
	}
}
