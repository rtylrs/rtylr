// Package csvsafe neutralizes spreadsheet formula injection in exported CSV.
package csvsafe

// Cell prefixes a quote to values a spreadsheet would evaluate as a formula.
func Cell(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

// Row applies Cell to every value, returning a new slice.
func Row(values ...string) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = Cell(v)
	}
	return out
}
