package application

import (
	"strings"
	"testing"
	"time"
)

func TestGeneralToolsForRequest(t *testing.T) {
	names := func(request string) string {
		var out []string
		for _, tool := range GeneralToolsForRequest(request) {
			out = append(out, tool.Name)
		}
		return strings.Join(out, ",")
	}
	for request, want := range map[string]string{
		"What's the weather in Beirut tomorrow?":                "weather.forecast",
		"convert 250 USD to LBP":                                "currency.convert",
		"Summarise https://example.com/post for me":             "web.fetch",
		"How many working days until 2026-12-24?":               "utilities.date_math",
		"hey how are you":                                       "",
		"what's the stock price of Apple and bitcoin?":          "markets.quote,markets.crypto",
		"What did we decide last time we talked about pricing?": "chats.search",
	} {
		if got := names(request); got != want {
			t.Errorf("%q -> %q, want %q", request, got, want)
		}
	}
}

func TestAssistantPromptOffersObviousGeneralToolsForSmallTalk(t *testing.T) {
	prompt := BuildAssistantPrompt("what's the weather like in Paris?", "", "", nil, time.Now())
	if !strings.Contains(prompt, "\n- weather.forecast [") {
		t.Fatal("weather.forecast was not offered for a weather question")
	}
}

func TestGeneralToolsForArabicAndFrenchRequests(t *testing.T) {
	for request, want := range map[string]string{
		"كيف الطقس في بيروت بكرا؟":      "weather.forecast",
		"حوّل ١٠٠ دولار لليرة":          "currency.convert",
		"Quelle météo à Paris demain ?": "weather.forecast",
	} {
		found := false
		for _, tool := range GeneralToolsForRequest(request) {
			found = found || tool.Name == want
		}
		if !found {
			t.Errorf("%q did not offer %s", request, want)
		}
	}
}
