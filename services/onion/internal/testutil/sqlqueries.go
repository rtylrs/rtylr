package testutil

import (
	"os"
	"strings"
	"testing"
)

// UnboundedMarker exempts a :many query from the bounded-result rule. Put it in
// a comment on or above the query with a reason, for example:
//
//	-- unbounded-ok: keyed by a caller supplied slice, so the caller bounds it.
//	-- name: ListThingsForOwners :many
//
// Only use it where the result set is bounded by something other than a LIMIT.
// A missing bound on a real list query is a defect, not something to mark.
const UnboundedMarker = "unbounded-ok:"

// SQLQuery is one sqlc query block read from a queries file. Text holds the
// statement together with the comment lines immediately above its name, so a
// marker written either place is visible to the checks below.
type SQLQuery struct {
	Name string
	Kind string
	Text string
}

// IsMany reports whether the query returns a row set rather than a single row.
func (q SQLQuery) IsMany() bool { return q.Kind == ":many" }

// upper is the statement folded to upper case, for keyword matching.
func (q SQLQuery) upper() string { return strings.ToUpper(q.Text) }

// HasLimit reports whether the statement bounds its result server side.
func (q SQLQuery) HasLimit() bool { return strings.Contains(q.upper(), "LIMIT") }

// IsPaginated reports whether the caller supplies the bound, which is what
// makes result ordering matter: without a total order the same row can appear
// on two pages or on none.
func (q SQLQuery) IsPaginated() bool { return strings.Contains(q.upper(), "LIMIT ?") }

// HasOrderBy reports whether the statement fixes a row order.
func (q SQLQuery) HasOrderBy() bool { return strings.Contains(q.upper(), "ORDER BY") }

// IsAggregate reports whether the statement groups rows.
//
// This does NOT imply a bounded result and must not be used to excuse a missing
// LIMIT. Grouping by a low cardinality column such as a status enum does bound
// the row count, but grouping by the row's own id, as ListGroups does to count
// members per group, returns one row per group and bounds nothing. Only an
// explicit UnboundedMarker with a reason distinguishes the two, so that is what
// the bounded-result rule requires.
func (q SQLQuery) IsAggregate() bool { return strings.Contains(q.upper(), "GROUP BY") }

// Exempt reports whether the query carries an explicit unbounded-ok marker.
func (q SQLQuery) Exempt() bool { return strings.Contains(q.Text, UnboundedMarker) }

// ParseSQLQueries splits a sqlc queries file into its named blocks.
//
// These checks deliberately read the .sql file rather than the generated Go: the
// invariants are properties of the statement text, and the file is the thing a
// contributor edits.
func ParseSQLQueries(t *testing.T, path string) []SQLQuery {
	t.Helper()

	// #nosec G304 -- Tests read only repository SQL fixture paths supplied by the test suite.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	const namePrefix = "-- name:"
	var (
		queries []SQLQuery
		lead    []string
		current *SQLQuery
	)
	flush := func() {
		if current != nil {
			queries = append(queries, *current)
			current = nil
		}
	}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, namePrefix) {
			flush()
			fields := strings.Fields(strings.TrimPrefix(trimmed, namePrefix))
			q := SQLQuery{Text: strings.Join(append(lead, line), "\n")}
			if len(fields) > 0 {
				q.Name = fields[0]
			}
			if len(fields) > 1 {
				q.Kind = fields[len(fields)-1]
			}
			current = &q
			lead = nil
			continue
		}
		if current != nil {
			current.Text += "\n" + line
		}
		// Track the run of comment lines directly above a name so a marker
		// written above the query, as bills.sql already does, still counts.
		if strings.HasPrefix(trimmed, "--") {
			lead = append(lead, line)
		} else {
			lead = nil
		}
	}
	flush()

	if len(queries) == 0 {
		t.Fatalf("no named queries found in %s", path)
	}
	return queries
}

// RequireBoundedListQueries asserts that every :many query bounds its result,
// and that every caller paginated query fixes a row order.
//
// This replaces counting how many statements match a literal. A count passes
// when someone adds an unbounded query, because the number of matches does not
// change, and fails when someone adds a compliant one. That is backwards, and
// it is how the drift in #892 reached main.
func RequireBoundedListQueries(t *testing.T, path string) {
	t.Helper()

	for _, q := range ParseSQLQueries(t, path) {
		if !q.IsMany() {
			continue
		}
		switch {
		case q.HasLimit(), q.Exempt():
			// Bounded server side, or exempted with a stated reason.
		default:
			t.Errorf("%s: %s returns many rows with no LIMIT. Bound it, or mark it with %q and a reason.",
				path, q.Name, UnboundedMarker)
		}
		if q.IsPaginated() && !q.HasOrderBy() {
			t.Errorf("%s: %s pages with LIMIT ? but has no ORDER BY, so rows can repeat or be skipped between pages.",
				path, q.Name)
		}
	}
}
