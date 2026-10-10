package application

import (
	"context"
	"strings"
)

// ToolRoutes maps a tool to the word sets of requests it served after
// Caret had to search for it.
type ToolRoutes map[string][][]string

// ToolRouter remembers which tool served which kind of request. The memory
// module implements it.
type ToolRouter interface {
	LearnToolRoute(ctx context.Context, userID int, tool string, words []string) error
	ToolRoutes(ctx context.Context, userID int) (map[string][][]string, error)
}

var routeStopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "was": true, "what": true, "who": true, "how": true,
	"can": true, "you": true, "your": true, "our": true, "with": true, "this": true, "that": true, "from": true,
	"have": true, "has": true, "show": true, "tell": true, "give": true, "please": true, "about": true, "any": true,
	"all": true, "get": true, "list": true, "need": true, "want": true, "there": true, "which": true, "many": true,
}

// RouteWords reduces a request to the words that identify what it asks for.
func RouteWords(request string) []string {
	seen := map[string]bool{}
	words := []string{}
	for _, word := range strings.Fields(normalizePromptToolText(request)) {
		word = strings.Trim(word, "'\"")
		if len([]rune(word)) < 3 || routeStopWords[word] || seen[word] {
			continue
		}
		seen[word] = true
		words = append(words, word)
	}
	return words
}

const maxRoutedTools = 3

// RoutedToolsForRequest offers the tools that served similar requests
// before: a route matches when most of its words appear in this request.
func RoutedToolsForRequest(request string, routes ToolRoutes) []ToolDef {
	if len(routes) == 0 {
		return nil
	}
	have := map[string]bool{}
	for _, word := range RouteWords(request) {
		have[word] = true
	}
	out := []ToolDef{}
	for _, def := range CoreTools {
		for _, words := range routes[def.Name] {
			matched := 0
			for _, word := range words {
				if have[word] {
					matched++
				}
			}
			if len(words) > 0 && matched >= 2 && matched*10 >= len(words)*6 || len(words) == 1 && matched == 1 {
				if tool, ok := ToolByName(def.Name); ok && promptToolCanExecute(def.Name) {
					out = append(out, tool)
				}
				break
			}
		}
		if len(out) == maxRoutedTools {
			break
		}
	}
	return out
}

// RouteTarget is one finished tool call, as the route learner sees it.
type RouteTarget struct {
	Name    string
	Success bool
	Output  map[string]any
}

// LearnedRoutes returns the tools of a turn that were found with tools.search
// and then ran successfully: the routes worth remembering.
func LearnedRoutes(calls []RouteTarget) []string {
	found := map[string]bool{}
	learned := []string{}
	for _, call := range calls {
		if call.Name == ToolSearchName && call.Success {
			for _, name := range searchedToolNames(call.Output) {
				found[name] = true
			}
			continue
		}
		if call.Success && found[call.Name] {
			learned = append(learned, call.Name)
			delete(found, call.Name)
		}
	}
	return learned
}

func searchedToolNames(output map[string]any) []string {
	names := []string{}
	switch tools := output["tools"].(type) {
	case []map[string]string:
		for _, tool := range tools {
			names = append(names, tool["name"])
		}
	case []any:
		for _, item := range tools {
			if tool, ok := item.(map[string]any); ok {
				if name, ok := tool["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
	}
	return names
}
