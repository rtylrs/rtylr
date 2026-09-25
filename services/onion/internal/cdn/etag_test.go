package cdn

import "testing"

func TestQuoteETag(t *testing.T) {
	if got := quoteETag("abc"); got != `"abc"` {
		t.Errorf("quoteETag = %q, want %q", got, `"abc"`)
	}
}

func TestEtagMatches(t *testing.T) {
	cases := []struct {
		ifNoneMatch string
		etag        string
		want        bool
	}{
		{"", `"a"`, false},
		{"*", `"a"`, true},
		{`"a"`, `"a"`, true},
		{`"b"`, `"a"`, false},
		{`"x", "a", "y"`, `"a"`, true},
		{`W/"a"`, `"a"`, true},
		{`  "a"  `, `"a"`, true},
	}
	for _, tc := range cases {
		if got := etagMatches(tc.ifNoneMatch, tc.etag); got != tc.want {
			t.Errorf("etagMatches(%q, %q) = %v, want %v", tc.ifNoneMatch, tc.etag, got, tc.want)
		}
	}
}
