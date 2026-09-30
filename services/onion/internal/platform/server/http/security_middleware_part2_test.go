package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func requestFor(t *testing.T, rawQuery string) *http.Request {
	t.Helper()
	return httptest.NewRequest(http.MethodGet, "/erp/contact/v1/contacts?"+rawQuery, http.NoBody)
}

func TestHasSuspiciousParams_AllowsOrdinaryFilterNames(t *testing.T) {
	// Regression: names were matched as substrings of keyword fragments, so
	// "created_at" hit "create", "updated_at" hit "update", "deleted_at" hit
	// "delete" and "description" hit "script". These returned 400 in
	// production for legitimate CRM/Finance/POS list requests.
	names := []string{
		"search", "description", "created_at", "updated_at", "deleted_at",
		"createdAt", "updatedAt", "deletedAt", "published_at", "archived_at",
		"selectAll", "selected", "selection", "aspect_ratio", "sharpen",
		"threshold", "hash", "page", "limit", "offset", "sort", "order_by",
		"status", "filter", "name", "email", "code", "signature", "state",
	}

	for _, name := range names {
		if hasSuspiciousParams(requestFor(t, name+"=1")) {
			t.Errorf("param name %q must not be blocked", name)
		}
	}
}

func TestHasSuspiciousParams_BlocksDenylistedNames(t *testing.T) {
	// Exact-match names only; the request must not be able to reach the sink.
	names := []string{
		"cmd", "CMD", "%20exec%20", "command", "system", "shell", "popen",
		"sql", "SQL", "raw_sql", "sql_query", "order_by_raw", "where_raw",
		"script", "javascript", "vbscript", "raw_template",
		"__proto__", "constructor",
		// Bracket / whitespace evasions fold to the same name.
		"[cmd]", "%5Bcmd%5D", "cmd[]",
	}

	for _, name := range names {
		if !hasSuspiciousParams(requestFor(t, name+"=x")) {
			t.Errorf("param name %q must be blocked", name)
		}
	}
}

func TestHasSuspiciousParams_BlocksMaliciousValues(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"sqli tautology", "q=%27+OR+1%3D1--"},
		{"sqli tautology lowercase", "q=%27+or+1%3D1--"},
		{"sqli union select", "q=1+UNION+SELECT+password+FROM+users"},
		{"sqli stacked", "q=1%3BDELETE+FROM+users"},
		{"sqli drop table", "q=%27%3B+DROP+TABLE+users+--"},
		{"sqli time based", "q=1+AND+SLEEP(5)"},
		{"sqli quote comment", "q=admin%27--"},
		{"crlf percent encoded", "q=alice%0d%0aSet-Cookie:%20a=b"},
		{"crlf raw", "q=alice%0a%0dLocation:%20https://evil.example"},
		{"nul byte", "q=alice%00.png"},
		{"path traversal", "file=../../etc/passwd"},
		{"path traversal backslash", "file=..%5c..%5cwindows"},
		{"path traversal encoded", "file=%2e%2e%2f%2e%2e%2fetc"},
		{"double encoded traversal", "file=%252e%252e%252fetc"},
		{"double encoded crlf", "q=%250d%250aSet-Cookie:%20a=b"},
	}

	for _, tc := range cases {
		if !hasSuspiciousParams(requestFor(t, tc.query)) {
			t.Errorf("%s: query %q must be blocked", tc.name, tc.query)
		}
	}
}

func TestHasSuspiciousParams_AllowsOrdinaryValues(t *testing.T) {
	// Values actually used by the CRM/Finance/POS clients and the Go handlers
	// (URL.Query().Get("url"|"code"|"signature"|"state"|"ticket"|...)).
	cases := []string{
		"q=alice",
		"name=O%27Brien", // apostrophe in a surname
		"description=please+delete+from+my+cart+and+retry", // prose, not SQL
		"url=https%3A%2F%2Fapp.example.com%2Fcb%3Fstate%3Dabc",
		"code=4%2F0cABCD...&state=xyz",
		"signature=4d0f2a1b",
		"q=C%2B%2B+and+R%23",
		"page=2&limit=50&offset=100",
		"file=2026-01-invoice.pdf",
		"search=caf%C3%A9+%D9%85%D8%B1%D8%AD%D8%A8%D8%A7", // Arabic / accented text
		"q=a.b.c...and..dots",
		"tags=a,b,c",
	}

	for _, query := range cases {
		if hasSuspiciousParams(requestFor(t, query)) {
			t.Errorf("query %q must not be blocked", query)
		}
	}
}

func TestHasSuspiciousParams_LongValueIsBounded(t *testing.T) {
	// 1 MiB single value and 64 KiB spread over many params. Both must be
	// decided quickly and deterministically: over-budget requests are rejected,
	// not silently truncated, so a payload cannot be hidden past the cut.
	huge := strings.Repeat("A", 1<<20)

	start := time.Now()
	if !hasSuspiciousParams(requestFor(t, "q="+huge)) {
		t.Error("an oversized parameter value must be rejected, not truncated past the check")
	}
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("scanning a 1 MiB value took %s; cost is not bounded", elapsed)
	}

	bulk := strings.Repeat("x=1&", 40000) // 40,000 pairs: past Go's 10,000-pair parse limit
	start = time.Now()
	if !hasSuspiciousParams(requestFor(t, bulk)) {
		t.Error("padding the query past the parse limit must not disable the check")
	}
	if total := time.Since(start); total > time.Second {
		t.Errorf("deciding an oversized query took %s; cost is not bounded", total)
	}

	// A long value that stays under the cap is still inspected in full.
	under := strings.Repeat("A", maxParamValueBytes-1) + "%27%20OR%201%3D1--"
	if !hasSuspiciousParams(requestFor(t, "q="+under)) {
		t.Error("a payload at the end of a long-but-legal value must still be caught")
	}
}

func TestHasSuspiciousParams_MalformedQueryIsNotAFalsePositive(t *testing.T) {
	// An unencoded "%" is malformed, not malicious. It must not become a 400.
	for _, query := range []string{"q=100%", "q=a%zz", "a=1;b=2"} {
		if hasSuspiciousParams(requestFor(t, query)) {
			t.Errorf("malformed but harmless query %q must not be blocked", query)
		}
	}
}

func TestHasSuspiciousParams_QueryPaddingCannotDisableTheCheck(t *testing.T) {
	// r.URL.Query() drops the whole map when Go's parser gives up, which made
	// every check here a no-op for a request padded past the pair limit.
	padded := "q=" + strings.Repeat("%27+OR+1%3D1--&pad", 2000)

	if !hasSuspiciousParams(requestFor(t, padded)) {
		t.Error("a padded query carrying an injection must still be blocked")
	}
}

func TestSecurityMiddleware_LegitimateFilterRequestsReachTheHandler(t *testing.T) {
	// End-to-end proof for the production regression: these list endpoints
	// returned 400 from this middleware before the name denylist was fixed.
	targets := []string{
		"/crm/contact/v1/contacts?search=alice",
		"/crm/contact/v1/contacts?q=alice&page=2&limit=50",
		"/finance/invoice/v1/invoices?sort=created_at&order_by=desc",
		"/finance/invoice/v1/invoices?description=Q1%20subscription",
		"/pos/order/v1/orders?created_at=2026-01-01",
		"/pos/order/v1/orders?updated_at=2026-01-01&deleted_at=2026-02-01",
		"/erp/employee/v1/employees?search=ali&updatedAt=2026-01-01",
	}

	for _, target := range targets {
		nextCalled := false
		handler := SecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, http.NoBody))

		if rec.Code != http.StatusOK || !nextCalled {
			t.Errorf("GET %s: status = %d, nextCalled = %t; legitimate list request must reach the handler",
				target, rec.Code, nextCalled)
		}
	}
}

func TestSecurityMiddleware_DoesNotEchoParamInBlockResponse(t *testing.T) {
	handler := SecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet,
		"/erp/contact/v1/contacts?q=%27+OR+1%3D1--", http.NoBody))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	body := rec.Body.String()
	for _, leak := range []string{"OR 1=1", "q=", "suspicious"} {
		if strings.Contains(body, leak) {
			t.Errorf("block response leaks %q: %q", leak, body)
		}
	}
}

func TestSecurityMiddleware_AllowListedPathNoLongerBypassesParamCheck(t *testing.T) {
	// The old code skipped the parameter check entirely on allow-listed paths,
	// so ?cmd=... reached the handler on /images/... once the name denylist
	// stopped being substring-based.
	for _, target := range []string{
		"/images/4c85c397-1234-9d1ef64f5e8b?cmd=whoami",
		"/console/stats?sql=1",
		"/upload/plan?script=alert(1)",
	} {
		nextCalled := false
		handler := SecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, http.NoBody))

		if rec.Code != http.StatusBadRequest || nextCalled {
			t.Errorf("GET %s: status = %d, nextCalled = %t; allow-listed paths must not bypass the param check",
				target, rec.Code, nextCalled)
		}
	}
}
