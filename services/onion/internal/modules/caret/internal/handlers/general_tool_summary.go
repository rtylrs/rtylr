package handlers

import (
	"fmt"
	"strings"
)

// generalToolSummary describes a general (non-business) tool result in one
// line for its card. ok is false when the tool has no special wording.
func generalToolSummary(name string, input, output map[string]any) (string, bool) {
	text := func(key string) string { value, _ := stringValue(output[key]); return value }
	count := func(key string) int { items, _ := collectionValue(output[key]); return len(items) }
	switch name {
	case "web.search":
		query, _ := stringValue(input["query"])
		return fmt.Sprintf("Found %s for %q.", plural(count("sources"), "source"), query), true
	case "web.fetch":
		return "Read " + firstNonEmpty(text("title"), text("url")) + ".", true
	case "agents.run":
		reports, _ := collectionValue(output["reports"])
		done := 0
		for _, report := range reports {
			if m, ok := report.(map[string]any); ok && m["status"] == "done" {
				done++
			}
		}
		return fmt.Sprintf("%d of %d sub-agents reported back, citing %s.", done, len(reports), plural(count("sources"), "source")), true
	case "weather.forecast":
		current, _ := output["current"].(map[string]any)
		conditions, _ := stringValue(current["conditions"])
		temperature, _ := stringValue(current["temperature_2m"])
		place := firstNonEmpty(text("location"), "the location")
		if conditions == "" {
			return "Checked the forecast for " + place + ".", true
		}
		return fmt.Sprintf("%s, %s%s in %s now.", capitalize(conditions), temperature, text("units"), place), true
	case "places.geocode":
		places, _ := collectionValue(output["places"])
		if len(places) == 0 {
			return "No place found.", true
		}
		return "Found " + itemDisplayName(places[0]) + ".", true
	case "currency.convert":
		conversions, _ := collectionValue(output["conversions"])
		parts := []string{}
		for _, item := range conversions {
			if m, ok := item.(map[string]any); ok {
				amount, _ := stringValue(m["amount"])
				currency, _ := stringValue(m["currency"])
				parts = append(parts, amount+" "+currency)
			}
		}
		amount, _ := stringValue(output["amount"])
		return fmt.Sprintf("%s %s = %s.", amount, text("from"), strings.Join(parts, ", ")), true
	case "calendar.public_holidays":
		return fmt.Sprintf("Found %s in %s for %s.", plural(count("holidays"), "public holiday"), text("country_code"), text("year")), true
	case "knowledge.wikipedia":
		return "Read the Wikipedia article on " + text("title") + ".", true
	case "knowledge.define":
		return "Looked up \"" + text("word") + "\".", true
	case "knowledge.country":
		return "Looked up " + text("name") + ".", true
	case "business.find":
		query, _ := stringValue(input["query"])
		return fmt.Sprintf("Found %s for %q.", plural(count("matches"), "record"), query), true
	case "chats.search":
		return fmt.Sprintf("Found %s in past chats.", plural(count("matches"), "mention")), true
	case "chats.read":
		return "Read the chat \"" + text("title") + "\".", true
	case "memory.list":
		return fmt.Sprintf("Caret remembers %s.", plural(count("memories"), "thing")), true
	case "tools.search":
		tools, _ := collectionValue(output["tools"])
		names := []string{}
		for _, tool := range tools {
			if m, ok := tool.(map[string]any); ok {
				if name, ok := stringValue(m["name"]); ok {
					names = append(names, name)
				}
			} else if m, ok := tool.(map[string]string); ok {
				names = append(names, m["name"])
			}
		}
		if len(names) == 0 {
			return "No matching tool.", true
		}
		if len(names) > 3 {
			names = names[:3]
		}
		return "Found " + strings.Join(names, ", ") + ".", true
	}
	if strings.HasPrefix(name, "utilities.") {
		if result, ok := stringValue(output["result"]); ok && result != "" {
			return "Result: " + result + ".", true
		}
		return "Worked it out.", true
	}
	return "", false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func capitalize(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
