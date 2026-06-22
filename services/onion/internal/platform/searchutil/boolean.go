package searchutil

import (
	"strings"
	"unicode"
)

// BooleanPrefixQuery turns user text into an indexed MySQL full-text boolean
// query where every token must match as a prefix.
func BooleanPrefixQuery(search string) string {
	terms := strings.FieldsFunc(strings.ToLower(search), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(terms) == 0 {
		return ""
	}

	out := make([]string, 0, len(terms))
	for _, term := range terms {
		if term == "" {
			continue
		}
		out = append(out, "+"+term+"*")
	}

	return strings.Join(out, " ")
}
