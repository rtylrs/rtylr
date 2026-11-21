package domain

import "testing"

func TestMatchWordKeysAreDistinctLowerCaseLetters(t *testing.T) {
	seen := map[string]bool{}
	for _, key := range MatchWords {
		if key == "" || len(key) > 13 || seen[key] {
			t.Fatalf("key %q is empty, too long for the stored word with a number, or repeated", key)
		}
		for _, r := range key {
			if r < 'a' || r > 'z' {
				t.Fatalf("key %q has %q: a digit would read as its number", key, r)
			}
		}
		seen[key] = true
	}
}

func TestParseMatchWord(t *testing.T) {
	for _, tc := range []struct {
		code   string
		key    string
		number int
		ok     bool
	}{
		{"panda", "panda", 0, true},
		{"panda2", "panda", 2, true},
		{"strawberry13", "strawberry", 13, true},
		{"panda1", "", 0, false},
		{"panda02", "", 0, false},
		{"pandas", "", 0, false},
		{"A7", "", 0, false},
		{"", "", 0, false},
	} {
		key, number, ok := ParseMatchWord(tc.code)
		if key != tc.key || number != tc.number || ok != tc.ok {
			t.Errorf("ParseMatchWord(%q) = %q, %d, %v; want %q, %d, %v", tc.code, key, number, ok, tc.key, tc.number, tc.ok)
		}
		if tc.ok && MatchWordCode(key, number) != tc.code {
			t.Errorf("MatchWordCode(%q, %d) = %q, want %q", key, number, MatchWordCode(key, number), tc.code)
		}
	}
}

func TestMatchColorOfAWordFollowsItsKeyAndNumber(t *testing.T) {
	seen := map[int]bool{}
	for _, key := range MatchWords {
		color := MatchColor(key)
		if color < 0 || color >= MatchColors {
			t.Fatalf("MatchColor(%s) = %d", key, color)
		}
		seen[color] = true
	}
	if len(seen) != MatchColors {
		t.Fatalf("words use %d of %d colours", len(seen), MatchColors)
	}
	if MatchColor("panda2") == MatchColor("panda") {
		t.Fatal("panda and panda 2 share a colour")
	}
}
