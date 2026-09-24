package segmentrules

import "testing"

func TestFilter_Empty(t *testing.T) {
	if frag, args := Filter(nil); frag != "" || len(args) != 0 {
		t.Fatalf("expected empty fragment, got %q %v", frag, args)
	}
	if frag, _ := Filter([]byte(`{}`)); frag != "" {
		t.Fatalf("expected empty fragment for {}, got %q", frag)
	}
}

func TestFilter_NumericRules(t *testing.T) {
	frag, args := Filter([]byte(`{"rules":["totalSpent > 500","visits >= 5"]}`))
	want := "total_spent > ? AND visit_count >= ?"
	if frag != want {
		t.Fatalf("frag = %q, want %q", frag, want)
	}
	if len(args) != 2 || args[0] != "500" || args[1] != "5" {
		t.Fatalf("args = %v", args)
	}
}

func TestFilter_RejectsUnknownFieldAndInjection(t *testing.T) {
	// Unknown fields and non-numeric values for numeric columns are dropped.
	frag, args := Filter([]byte(`{"rules":["password = x","visits > drop table"]}`))
	if frag != "" || len(args) != 0 {
		t.Fatalf("expected nothing to survive, got %q %v", frag, args)
	}
}

func TestFilter_StringContainsEscapesWildcards(t *testing.T) {
	frag, args := Filter([]byte(`{"rules":["country contains US%50"]}`))
	want := "country LIKE ? ESCAPE '!'"
	if frag != want {
		t.Fatalf("frag = %q, want %q", frag, want)
	}
	if len(args) != 1 || args[0] != "%US!%50%" {
		t.Fatalf("args = %v", args)
	}
}

func TestFilter_TagsJSONContains(t *testing.T) {
	frag, args := Filter([]byte(`{"rules":["tags contains vip"]}`))
	if frag != "JSON_CONTAINS(tags, ?)" {
		t.Fatalf("frag = %q", frag)
	}
	if len(args) != 1 || args[0] != `"vip"` {
		t.Fatalf("args = %v", args)
	}
}

func TestFilter_Boolean(t *testing.T) {
	frag, args := Filter([]byte(`{"rules":["isActive = true"]}`))
	if frag != "is_active = ?" || len(args) != 1 || args[0] != true {
		t.Fatalf("frag=%q args=%v", frag, args)
	}

	frag, args = Filter([]byte(`{"rules":["isActive = false"]}`))
	if frag != "is_active = ?" || len(args) != 1 || args[0] != false {
		t.Fatalf("false rule frag=%q args=%v", frag, args)
	}
}
