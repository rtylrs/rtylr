package planner

import (
	"context"
	"math"
	"strings"

	"github.com/voxire/onion/internal/ai/classifier/model"
)

// LexicalScorer is the deterministic fallback used when no classifier checkpoint
// is configured. It scores one example at a time: the planner falls back to
// per-field scoring when a scorer has no batch path.
type LexicalScorer struct{}

const lexicalTemperature = 0.35

// Score implements Scorer.
func (LexicalScorer) Score(ctx context.Context, ex model.Example) ([]float64, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	label := elementLabel(ex.Context)
	scores := make([]float64, len(ex.Options))
	for i, option := range ex.Options {
		scores[i] = lexicalScore(label, option)
	}
	best := 0
	for i, s := range scores {
		if s > scores[best] {
			best = i
		}
	}
	probs := make([]float64, len(scores))
	var sum float64
	for i, s := range scores {
		e := math.Exp(s / lexicalTemperature)
		probs[i] = e
		sum += e
	}
	for i := range probs {
		probs[i] /= sum
	}
	return probs, best, nil
}

func lexicalScore(fieldLabel, option string) float64 {
	switch option {
	case ActionCheck:
		return 1.2 * booleanFieldPrior(fieldLabel)
	case ActionClick:
		if isSubmitLabel(fieldLabel) {
			return 1.2
		}
		return 0.05
	case ActionSkip:
		return 0.4
	}
	if !strings.HasPrefix(option, "fill ") {
		return 0
	}
	entityLabel, _, found := strings.Cut(strings.TrimPrefix(option, "fill "), ":")
	if !found {
		return 0
	}
	return labelSimilarity(fieldLabel, entityLabel)
}

func labelSimilarity(a, b string) float64 {
	na, nb := normalizePromptLabel(a), normalizePromptLabel(b)
	if na == "" || nb == "" {
		return 0
	}
	if na == nb {
		return 1.6
	}
	if strings.Contains(na, nb) || strings.Contains(nb, na) {
		return 1.0
	}
	tokensA := strings.Fields(na)
	tokensB := strings.Fields(nb)
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}
	shared := 0
	for _, ta := range tokensA {
		for _, tb := range tokensB {
			if ta == tb {
				shared++
				break
			}
		}
	}
	if shared == 0 {
		return 0
	}
	overlap := float64(shared) / float64(len(tokensA))
	return 0.7 * overlap
}

func booleanFieldPrior(label string) float64 {
	n := normalizePromptLabel(label)
	for _, token := range strings.Fields(n) {
		switch token {
		case "yes", "no", "has", "is", "active", "agree", "agreed", "consent", "subscribe",
			"enable", "enabled", "include", "applicable", "checked", "unchecked", "already", "same":
			return 1
		}
	}
	for _, phrase := range []string{"opt in", "opt out"} {
		if strings.Contains(n, phrase) {
			return 1
		}
	}
	return 0.15
}

func isSubmitLabel(label string) bool {
	return normalizeLabel(label) == "submit" || normalizeLabel(label) == "submit form"
}

func normalizePromptLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("_", " ", "-", " ", ".", " ", ",", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func elementLabel(context string) string {
	for _, line := range strings.Split(context, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ELEMENT ") {
			continue
		}
		rest := strings.TrimPrefix(line, "ELEMENT ")
		if start := strings.Index(rest, ` "`); start >= 0 {
			remainder := rest[start+2:]
			if end := strings.Index(remainder, `"`); end > 0 {
				return remainder[:end]
			}
		}
		if _, after, found := strings.Cut(rest, " "); found {
			label, _, _ := strings.Cut(after, ` value=`)
			return strings.TrimSpace(label)
		}
		return strings.TrimSpace(rest)
	}
	return ""
}
