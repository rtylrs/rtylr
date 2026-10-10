package application

// generalToolDefs are the tools that need no business data. Each has a
// handler registered in generalTools by its own file. Inputs and outputs are
// spelled out here because the generic contract hints only fit business
// list/get tools.
var generalToolDefs = []ToolDef{
	{
		Name:        "web.fetch",
		Description: "Read a public web page and return its title and readable text. Use it after web.search when snippets are not enough, or when the person shares a link.",
		Module:      "web", Scope: "read",
		InputHint:  `{"url":"https://...","query":"optional: what you are looking for, so the relevant part of a long page is kept"}`,
		OutputHint: "The final URL, page title, description and up to 12,000 characters of text; long pages are trimmed around the query.",
		Freshness:  "live at request time", SideEffect: "read-only; fetches one public page",
	},
	{
		Name:        "places.geocode",
		Description: "Look up a city or place by name: coordinates, country, timezone and population. Add a region after a comma to disambiguate, e.g. \"Paris, Texas\".",
		Module:      "places", Scope: "read",
		InputHint:  `{"name":"city or place"}`,
		OutputHint: "Up to five matching places with coordinates, country, IANA timezone and population.",
		Freshness:  "reference data", SideEffect: "read-only",
	},
	{
		Name:        "weather.forecast",
		Description: "Current weather and a daily forecast (up to 16 days) for a place: conditions, highs and lows, rain chance, wind, sunrise and sunset.",
		Module:      "weather", Scope: "read",
		InputHint:  `{"location":"city or place","days":3,"units":"celsius|fahrenheit"} or {"latitude":0,"longitude":0}`,
		OutputHint: "Current conditions and one entry per day, in the place's local timezone.",
		Freshness:  "live forecast", SideEffect: "read-only",
	},
	{
		Name:        "currency.convert",
		Description: "Convert money between currencies with today's exchange rates (about 160 currencies, including LBP, AED, SAR) or the rates on a past date.",
		Module:      "currency", Scope: "read",
		InputHint:  `{"amount":100,"from":"USD","to":"EUR,GBP","date":"optional YYYY-MM-DD"}`,
		OutputHint: "The rate and converted amount per target currency, the rate date and its source.",
		Freshness:  "daily rates", SideEffect: "read-only",
	},
	{
		Name:        "calendar.public_holidays",
		Description: "List a country's public holidays for a year, with dates and weekdays. Useful for scheduling, staffing and delivery estimates.",
		Module:      "calendar", Scope: "read",
		InputHint:  `{"country_code":"AE","year":2026,"upcoming":true}`,
		OutputHint: "Each holiday's date, weekday, English and local name, and whether it is regional.",
		Freshness:  "reference data", SideEffect: "read-only",
	},
}
