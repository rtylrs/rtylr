package application

import (
	"regexp"
	"strings"
)

// generalToolTriggers are phrases that make a general tool obviously useful.
// These tools are offered whatever the request is about, because "what's the
// weather" or "convert 50 USD" is small talk to the business classifier.
var generalToolTriggers = map[string][]string{
	"weather.forecast":         {"weather", "forecast", "rain", "temperature", "sunny", "humid", "snow", "storm", "طقس", "الطقس", "مطر", "حرارة", "météo", "pluie"},
	"weather.history":          {"weather last", "weather on", "was it raining", "rained", "weather yesterday"},
	"weather.air_quality":      {"air quality", "aqi", "pollution", "smog"},
	"currency.convert":         {"exchange rate", "convert", "in dollars", "in euros", "usd to", "eur to", "to usd", "to lbp", "lbp", "currency", "سعر الصرف", "دولار", "يورو", "ليرة", "صرف", "taux de change", "convertir"},
	"markets.quote":            {"stock price", "share price", "ticker", "nasdaq", "s&p", "dow jones", "stock market", "shares of", "سهم", "أسهم", "بورصة", "action en bourse"},
	"markets.crypto":           {"bitcoin", "btc", "ethereum", "crypto", "usdt", "solana", "بيتكوين", "عملة رقمية"},
	"calendar.public_holidays": {"public holiday", "bank holiday", "holidays in", "national holiday", "eid", "عطلة", "عيد", "عطل رسمية", "jour férié", "jours fériés"},
	"knowledge.wikipedia":      {"wikipedia", "who was", "who is", "history of", "tell me about"},
	"knowledge.define":         {"define ", "definition of", "meaning of", "what does the word", "synonym", "معنى كلمة", "ما معنى", "que veut dire"},
	"knowledge.country":        {"capital of", "population of", "currency of", "dialing code", "calling code"},
	"knowledge.papers":         {"research on", "studies on", "study on", "papers on", "scientific", "evidence that", "peer reviewed"},
	"places.search":            {"near me", "nearby", " near ", "competitors in", "cafes in", "restaurants in", "بالقرب", "قريب من", "près de"},
	"utilities.date_math":      {"days until", "days between", "how many days", "working days", "business days", "weeks until", "days from now", "days ago", "كم يوم", "combien de jours"},
	"utilities.convert_time":   {"time in", "timezone", "time zone", "what time is it in", "pm in", "am in", "كم الساعة", "الساعة في", "quelle heure"},
	"utilities.convert_units":  {"km to", "miles to", "kg to", "lbs to", "celsius", "fahrenheit", "liters to", "gallons", "square meters", "sqft", "dunam"},
	"utilities.statistics":     {"average of", "median", "standard deviation", "mean of"},
	"utilities.business_math":  {"margin", "markup", "vat", "break even", "break-even", "loan", "interest", "cagr", "discount"},
	"utilities.random":         {"pick a winner", "random winner", "giveaway", "raffle", "draw a name", "pick randomly", "shuffle", "random number", "سحب", "اختر عشوائيا"},
	"text.count":               {"word count", "how many words", "how many characters", "character count", "too long for"},
	"chats.search":             {"last time we", "we discussed", "we talked about", "i told you", "i mentioned", "earlier chat", "we decided"},
	"memory.list":              {"what do you know about me", "what do you remember", "your memory", "my memories"},
	"buddy.routines":           {"routine", "check-in", "check in", "remind me", "every day", "every morning", "every night", "every week", "every monday", "each day", "daily", "weekly", "pause", "stop asking", "كل يوم", "ذكرني", "tous les jours", "rappelle-moi"},
	"buddy.create_routine":     {"routine", "check-in", "check in", "remind me", "every day", "every morning", "every night", "every week", "every monday", "each day", "daily", "weekly", "كل يوم", "ذكرني", "tous les jours", "rappelle-moi"},
}

var urlInText = regexp.MustCompile(`https?://[^\s]+|\b[a-z0-9-]+\.(com|org|net|io|co|ai|dev|app|lb|ae|sa)(/[^\s]*)?\b`)

const maxGeneralShortlist = 4

// GeneralToolsForRequest returns the general tools a request obviously
// needs, in registry order.
func GeneralToolsForRequest(request string) []ToolDef {
	text := " " + strings.ToLower(request) + " "
	out := []ToolDef{}
	add := func(name string) {
		for _, tool := range out {
			if tool.Name == name {
				return
			}
		}
		if tool, ok := ToolByName(name); ok && len(out) < maxGeneralShortlist {
			out = append(out, tool)
		}
	}
	if urlInText.MatchString(text) {
		add("web.fetch")
	}
	for _, def := range generalToolDefs {
		for _, trigger := range generalToolTriggers[def.Name] {
			if strings.Contains(text, trigger) {
				add(def.Name)
				break
			}
		}
	}
	for _, name := range []string{"chats.search", "memory.list", "buddy.routines", "buddy.create_routine"} {
		for _, trigger := range generalToolTriggers[name] {
			if strings.Contains(text, trigger) {
				add(name)
				break
			}
		}
	}
	return out
}
