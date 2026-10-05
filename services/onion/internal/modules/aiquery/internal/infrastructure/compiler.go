package infrastructure

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"regexp"
	"strings"
)

var (
	allowedAggregates = set("sum", "count", "avg", "min", "max")
	allowedOperators  = set("=", "!=", "<", "<=", ">", ">=", "like", "in")
	aliasRe           = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// BuildSQL validates the spec against the allowlist and compiles it to a
// parameterized SELECT that is always constrained to businessID.
//
// businessID must be the id resolved by tenant.ResolveBusinessID. owner_id is
// deliberately not a substitute: it is shared by every business of an account,
// so scoping by it returns rows from businesses the caller has no membership in.
func BuildSQL(spec domain.QuerySpec, businessID int64) (string, []any, error) {
	if businessID <= 0 {
		return "", nil, domain.ErrNoBusinessScope
	}

	columns, ok := domain.Schema()[spec.View]
	cols := set(columns...)
	if !ok {
		return "", nil, fmt.Errorf("%w: %q", domain.ErrUnknownView, spec.View)
	}

	selects, err := buildSelects(spec, cols)
	if err != nil {
		return "", nil, err
	}

	// business_id is always bound — the core isolation guarantee.
	b := sq.Select(selects...).From(spec.View).Where(sq.Eq{"business_id": businessID})

	for _, f := range spec.Where {
		pred, perr := filterPredicate(f, cols)
		if perr != nil {
			return "", nil, perr
		}
		b = b.Where(pred)
	}

	for _, g := range spec.GroupBy {
		if !allowed(cols, g) {
			return "", nil, fmt.Errorf("%w: group by %q", domain.ErrUnknownColumn, g)
		}
		b = b.GroupBy(g)
	}

	aliases := aggregationAliases(spec)
	for _, o := range spec.OrderBy {
		if !allowed(cols, o.Col) && !aliases[o.Col] {
			return "", nil, fmt.Errorf("%w: order by %q", domain.ErrUnknownColumn, o.Col)
		}
		dir := "ASC"
		if strings.EqualFold(o.Dir, "desc") {
			dir = "DESC"
		}
		b = b.OrderBy(o.Col + " " + dir)
	}

	limit := spec.Limit
	if limit <= 0 {
		limit = domain.DefaultLimit
	}
	if limit > domain.MaxRows {
		limit = domain.MaxRows
	}
	b = b.Limit(uint64(limit))

	return b.PlaceholderFormat(sq.Question).ToSql()
}

func buildSelects(spec domain.QuerySpec, cols map[string]struct{}) ([]string, error) {
	var selects []string

	for _, c := range spec.Columns {
		if !allowed(cols, c) {
			return nil, fmt.Errorf("%w: %q", domain.ErrUnknownColumn, c)
		}
		selects = append(selects, c)
	}

	for _, a := range spec.Aggregations {
		fn := strings.ToLower(strings.TrimSpace(a.Func))
		if !inSet(allowedAggregates, fn) {
			return nil, fmt.Errorf("%w: %q", domain.ErrBadAggregate, a.Func)
		}
		alias := sanitizeAlias(a.As)
		if fn == "count" && a.Col == "" {
			selects = append(selects, fmt.Sprintf("COUNT(*) AS %s", orDefault(alias, "count")))
			continue
		}
		if !allowed(cols, a.Col) {
			return nil, fmt.Errorf("%w: aggregate on %q", domain.ErrUnknownColumn, a.Col)
		}
		selects = append(selects, fmt.Sprintf("%s AS %s", aggregateExpr(fn, a.Col), orDefault(alias, fn+"_"+a.Col)))
	}

	if len(selects) == 0 {
		// Default to all allowed columns rather than erroring on a bare view.
		for c := range cols {
			selects = append(selects, c)
		}
		if len(selects) == 0 {
			return nil, domain.ErrEmptySelect
		}
	}

	return selects, nil
}

func aggregateExpr(fn, col string) string {
	switch fn {
	case "sum":
		return fmt.Sprintf("COALESCE(SUM(%s), 0)", col)
	default:
		return fmt.Sprintf("%s(%s)", strings.ToUpper(fn), col)
	}
}

func filterPredicate(f domain.Filter, cols map[string]struct{}) (sq.Sqlizer, error) {
	if !allowed(cols, f.Col) {
		return nil, fmt.Errorf("%w: filter on %q", domain.ErrUnknownColumn, f.Col)
	}
	op := strings.ToLower(strings.TrimSpace(f.Op))
	if !inSet(allowedOperators, op) {
		return nil, fmt.Errorf("%w: %q", domain.ErrBadOperator, f.Op)
	}
	if strings.TrimSpace(f.ValueColumn) != "" {
		valueCol := strings.TrimSpace(f.ValueColumn)
		if !allowed(cols, valueCol) {
			return nil, fmt.Errorf("%w: value column %q", domain.ErrUnknownColumn, valueCol)
		}
		if op == "in" || op == "like" {
			return nil, fmt.Errorf("%w: column comparison with %q", domain.ErrBadOperator, f.Op)
		}
		return sq.Expr(fmt.Sprintf("%s %s %s", strings.TrimSpace(f.Col), op, valueCol)), nil
	}

	switch op {
	case "=":
		return sq.Eq{f.Col: f.Value}, nil
	case "!=":
		return sq.NotEq{f.Col: f.Value}, nil
	case "<":
		return sq.Lt{f.Col: f.Value}, nil
	case "<=":
		return sq.LtOrEq{f.Col: f.Value}, nil
	case ">":
		return sq.Gt{f.Col: f.Value}, nil
	case ">=":
		return sq.GtOrEq{f.Col: f.Value}, nil
	case "like":
		return sq.Like{f.Col: f.Value}, nil
	case "in":
		return sq.Eq{f.Col: f.Value}, nil // squirrel renders a slice value as IN (...)
	}
	return nil, fmt.Errorf("%w: %q", domain.ErrBadOperator, f.Op)
}

func aggregationAliases(spec domain.QuerySpec) map[string]bool {
	out := map[string]bool{}
	for _, a := range spec.Aggregations {
		alias := sanitizeAlias(a.As)
		if alias == "" {
			if strings.ToLower(a.Func) == "count" && a.Col == "" {
				alias = "count"
			} else {
				alias = strings.ToLower(a.Func) + "_" + a.Col
			}
		}
		out[alias] = true
	}
	return out
}

func set(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, i := range items {
		m[i] = struct{}{}
	}
	return m
}

func inSet(m map[string]struct{}, k string) bool { _, ok := m[k]; return ok }
func allowed(cols map[string]struct{}, c string) bool {
	_, ok := cols[strings.TrimSpace(c)]
	return ok
}

func sanitizeAlias(s string) string { return aliasRe.ReplaceAllString(strings.TrimSpace(s), "") }
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
