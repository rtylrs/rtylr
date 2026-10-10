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
}
