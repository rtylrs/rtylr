package httprequest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPResponse_StatusBuckets(t *testing.T) {
	cases := []struct {
		code int
		ok   bool
		is4  bool
		is5  bool
	}{
		{200, true, false, false},
		{204, true, false, false},
		{299, true, false, false},
		{300, false, false, false},
		{400, false, true, false},
		{404, false, true, false},
		{499, false, true, false},
		{500, false, false, true},
		{599, false, false, true},
	}
	for _, tc := range cases {
		r := HTTPResponse{StatusCode: tc.code}
		if r.IsOk() != tc.ok || r.Is4XX() != tc.is4 || r.Is5XX() != tc.is5 {
			t.Errorf("status %d: ok=%v want %v, 4xx=%v want %v, 5xx=%v want %v",
				tc.code, r.IsOk(), tc.ok, r.Is4XX(), tc.is4, r.Is5XX(), tc.is5)
		}
	}
}

func TestHTTPResponse_WasRedirected(t *testing.T) {
	if r := (HTTPResponse{}); func() bool { ok, _ := r.WasRedirected(); return ok }() {
		t.Error("empty redirect URL should report false")
	}
	r := HTTPResponse{}
	r.finalRedirectURL = "http://final"
	ok, url := r.WasRedirected()
	if !ok || url != "http://final" {
		t.Errorf("expected (true, http://final), got (%v, %q)", ok, url)
	}
}

func TestGetReturnsBodyAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer srv.Close()

	resp, err := Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !resp.IsOk() {
		t.Errorf("expected ok status, got %d", resp.StatusCode)
	}
	if resp.Headers.Get("X-Test") != "yes" {
		t.Errorf("expected X-Test header to round-trip")
	}

	var payload struct {
		Hello string `json:"hello"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Hello != "world" {
		t.Errorf("expected hello=world, got %q", payload.Hello)
	}
}

func TestPostBodyIsForwarded(t *testing.T) {
	var got []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = make([]byte, r.ContentLength)
		_, _ = r.Body.Read(got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	resp, err := Post(context.Background(), srv.URL, []byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if string(got) != `{"a":1}` {
		t.Errorf("expected body forwarded, got %q", got)
	}
}

func TestReadResponseBodyRejectsOversizedBody(t *testing.T) {
	resp, err := readResponseBody(strings.NewReader("abcdef"), 5)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
	if string(resp) != "abcde" {
		t.Fatalf("expected capped body prefix, got %q", resp)
	}
}

func TestGetReturnsErrorWhenResponseExceedsLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer srv.Close()

	resp, err := Get(context.Background(), srv.URL, WithMaxResponseBytes(5))
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
	if string(resp.Body) != "abcde" {
		t.Fatalf("expected capped response body, got %q", resp.Body)
	}
}

func TestWithMaxResponseBytesAllowsLargerBody(t *testing.T) {
	const body = "abcdef"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	resp, err := Get(context.Background(), srv.URL, WithMaxResponseBytes(int64(len(body))))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if string(resp.Body) != body {
		t.Fatalf("expected full body, got %q", resp.Body)
	}
}
