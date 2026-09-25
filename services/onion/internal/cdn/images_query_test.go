package cdn

import (
	"net/http/httptest"
	"testing"
)

func TestQueryInt(t *testing.T) {
	cases := map[string]int{
		"/x?w=256": 256,
		"/x?w=0":   0,
		"/x":       0,
		"/x?w=":    0,
		"/x?w=abc": 0,
		"/x?w=-5":  -5,
	}
	for url, want := range cases {
		req := httptest.NewRequest("GET", url, nil)
		if got := queryInt(req.URL.Query(), "w"); got != want {
			t.Errorf("queryInt(%q) = %d, want %d", url, got, want)
		}
	}
}
