package application

import (
	"regexp"
	"strconv"
	"strings"
)

// Budget bands differ between the contact form ("$8K - $20K") and the quote
// form ("$5k-$15k"), and free text arrives too. The tier is decided by the
// largest figure in the string, in thousands of dollars.
const (
	tierHot  = "HOT"
	tierWarm = "WARM"
	tierNew  = "NEW"

	hotFromThousands  = 15
	warmFromThousands = 5
)

// budgetFigure matches "20K", "5k", "15k+" (K-suffixed) or "10,000", "25,000"
// (thousands-separated). Bare digit runs like "71234567" or "20260115" are
// ignored, and so is a "k" that starts a word ("15 key pages", "8 keywords"):
// the trailing \b requires the k to end a token rather than begin one.
var budgetFigure = regexp.MustCompile(`(?i)(\d{1,3}(?:,\d{3})+)|(\d+)\s*(k)\b`)

func budgetTier(budget string) string {
	maxThousands := 0.0
	for _, m := range budgetFigure.FindAllStringSubmatch(budget, -1) {
		if m[1] != "" {
			// "10,000" is thousands-separated dollars.
			raw := strings.ReplaceAll(m[1], ",", "")
			n, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				continue
			}
			if n/1000 > maxThousands {
				maxThousands = n / 1000
			}
			continue
		}
		if m[3] != "" {
			// "20K" is twenty thousand.
			n, err := strconv.ParseFloat(m[2], 64)
			if err != nil {
				continue
			}
			if n > maxThousands {
				maxThousands = n
			}
		}
		// Bare digits with no K and no comma separator are not budget figures
		// ("within 2 weeks", "71234567", phone numbers), so they are ignored.
	}
	switch {
	case maxThousands >= hotFromThousands:
		return tierHot
	case maxThousands >= warmFromThousands:
		return tierWarm
	default:
		return tierNew
	}
}
