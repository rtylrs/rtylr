package application

import (
	"regexp"
	"strings"
)

var numericIDPattern = regexp.MustCompile(`\b\d{2,}\b`)

func ClarificationHintForRequest(request, systemContext string) string {
	text := normalizePromptToolText(request)
	if text == "" {
		return ""
	}
	context := normalizePromptToolText(systemContext)

	if referencesEntity(text, "order") && !hasEntityContext(context, "order") && !numericIDPattern.MatchString(request) {
		return "The order is not identified. Ask which order they mean before checking or changing it."
	}
	if referencesEntity(text, "product") && !hasEntityContext(context, "product") && !numericIDPattern.MatchString(request) {
		return "The product is not identified. Ask which product they mean before checking or changing it."
	}
	if referencesEntity(text, "customer") && !hasEntityContext(context, "customer") && !numericIDPattern.MatchString(request) {
		return "The customer is not identified. Ask which customer they mean before checking or changing it."
	}
	if referencesEntity(text, "employee") && !hasEntityContext(context, "employee") && !numericIDPattern.MatchString(request) {
		return "The employee is not identified. Ask which team member they mean before checking or changing it."
	}
	if (containsPromptPhrase(text, "this branch") || containsPromptPhrase(text, "current branch") || containsPromptPhrase(text, "selected branch")) &&
		!hasAnyContext(context, []string{"branch id", "branch name", "location id", "location name", "selected branch", "selected location"}) {
		return "The branch is not identified. Ask whether to use all branches or a specific branch."
	}
	if hasAmbiguousPointer(text) && !hasAnyContext(context, []string{"selected", "selected ids", "visible rows", "current row", "current item"}) {
		return "The referenced item is not identified. Ask one short question to clarify what they mean."
	}
	return ""
}

func referencesEntity(text, entity string) bool {
	plural := entity + "s"
	if entity == "employee" {
		plural = "employees"
	}
	return (containsPromptPhrase(text, "this "+entity) ||
		containsPromptPhrase(text, "that "+entity) ||
		containsPromptPhrase(text, "current "+entity) ||
		containsPromptPhrase(text, "selected "+entity) ||
		containsPromptPhrase(text, "the "+entity) ||
		containsPromptPhrase(text, "these "+plural) ||
		containsPromptPhrase(text, "those "+plural) ||
		containsPromptPhrase(text, "selected "+plural))
}

func hasAmbiguousPointer(text string) bool {
	words := map[string]struct{}{}
	for _, word := range strings.Fields(text) {
		words[word] = struct{}{}
	}
	for _, pointer := range []string{"this", "that", "these", "those", "selected", "it"} {
		if _, ok := words[pointer]; ok {
			return true
		}
	}
	for _, pointer := range []string{"current item", "current row"} {
		if strings.Contains(text, pointer) {
			return true
		}
	}
	return false
}

func hasEntityContext(context, entity string) bool {
	if context == "" {
		return false
	}
	plural := entity + "s"
	if entity == "employee" {
		plural = "employees"
	}
	return hasAnyContext(context, []string{
		entity + " id",
		entity + " ids",
		plural + " id",
		plural + " ids",
		"selected " + entity,
		"selected " + plural,
		"current " + entity,
		"current " + plural,
	})
}

func hasAnyContext(context string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(context, normalizePromptToolText(needle)) {
			return true
		}
	}
	return false
}

func containsPromptPhrase(text, phrase string) bool {
	return strings.Contains(" "+text+" ", " "+normalizePromptToolText(phrase)+" ")
}
