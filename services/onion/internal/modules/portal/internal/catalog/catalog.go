// Package catalog is the map of the Portal: every page, what people do there,
// and the navigation they can arrange. The Portal's tests check it against
// the real routes and sidebars, so Caret never points at a page that isn't
// there.
package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/voxire/onion/internal/modules/portal/internal/layout"
)

//go:embed navigation.json
var navigationJSON []byte

//go:embed pages.json
var pagesJSON []byte

// Page is one Portal page.
type Page struct {
	Path    string   `json:"path"`
	App     string   `json:"app"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Tasks   []string `json:"tasks"`
	// Actions are the English labels of the page's main buttons.
	Actions  []string `json:"actions"`
	Keywords []string `json:"keywords"`
}

var apps = []string{"workspace", "insights", "flow", "crm", "erp", "finance", "hr", "recruit", "automations", "settings", "caret"}

var (
	navigation layout.Navigation
	pages      []Page
	pageTerms  [][4]map[string]bool
	// taskTerms holds each task and button of a page on its own.
	taskTerms [][]map[string]bool
)

func init() {
	if err := load(); err != nil {
		panic(fmt.Errorf("portal catalog: %w", err))
	}
}

func load() error {
	if err := json.Unmarshal(navigationJSON, &navigation); err != nil {
		return fmt.Errorf("navigation.json: %w", err)
	}
	for _, area := range []layout.Area{navigation.Rail, navigation.Workspace} {
		seen := map[string]bool{}
		for _, item := range area.Items {
			if item.ID == "" || item.Label == "" || seen[item.ID] {
				return fmt.Errorf("navigation.json: %s has an empty or repeated item %q", area.Name, item.ID)
			}
			seen[item.ID] = true
		}
	}
	if err := json.Unmarshal(pagesJSON, &pages); err != nil {
		return fmt.Errorf("pages.json: %w", err)
	}
	seen := map[string]bool{}
	for _, page := range pages {
		if !strings.HasPrefix(page.Path, "/") || seen[page.Path] {
			return fmt.Errorf("pages.json: path %q is not absolute or is repeated", page.Path)
		}
		if page.Title == "" || page.Summary == "" || !slices.Contains(apps, page.App) {
			return fmt.Errorf("pages.json: %s needs a title, a summary and a known app", page.Path)
		}
		seen[page.Path] = true
		tasks := []map[string]bool{}
		for _, task := range append(slices.Clone(page.Tasks), page.Actions...) {
			tasks = append(tasks, terms(task))
		}
		taskTerms = append(taskTerms, tasks)
		pageTerms = append(pageTerms, [4]map[string]bool{
			terms(page.Title),
			terms(strings.Join(page.Keywords, " ")),
			terms(strings.Join(append(slices.Clone(page.Tasks), page.Actions...), " ")),
			terms(page.Summary + " " + page.App),
		})
	}
	return nil
}

// Navigation is the default arrangement of every area people can arrange.
func Navigation() layout.Navigation { return navigation }

// Pages is every page, in Portal order.
func Pages() []Page { return pages }

// termWeights rank where a word matched: the title, the keywords, the tasks
// and buttons, then the summary.
var termWeights = [4]int{6, 4, 3, 1}

// Search returns the pages that best match a description such as "add team
// members" or "unpaid invoices", best first.
func Search(query string, limit int) []Page {
	words := terms(query)
	if len(words) == 0 {
		return nil
	}
	type scored struct {
		index, score int
	}
	var matches []scored
	for i := range pages {
		score, matched, generic := 0, 0, 0
		for word := range words {
			hit := 0
			for field, weight := range termWeights {
				if pageTerms[i][field][word] {
					hit += weight
				}
			}
			switch {
			case hit == 0:
			case genericWords[word]:
				generic += hit
			default:
				score += hit
				matched++
			}
		}
		if score+generic == 0 {
			continue
		}
		// A page that answers more of the words wins, and one whose whole
		// title was asked for ("payroll") beats its settings page. Verbs like
		// "add" only break ties: most pages add something.
		score = score*matched + generic/3
		if titled := pageTerms[i][0]; len(titled) > 0 && all(titled, words) {
			score += 8
		}
		// One task or button says the whole request ("Add a product").
		if len(words) > 1 && slices.ContainsFunc(taskTerms[i], func(task map[string]bool) bool { return all(words, task) }) {
			score += 20
		}
		matches = append(matches, scored{i, score})
	}
	sort.SliceStable(matches, func(a, b int) bool { return matches[a].score > matches[b].score })
	out := []Page{}
	for _, match := range matches {
		if len(out) == limit {
			break
		}
		out = append(out, pages[match.index])
	}
	return out
}

func all(want, have map[string]bool) bool {
	for word := range want {
		if !have[word] {
			return false
		}
	}
	return true
}

var genericWords = map[string]bool{
	"add": true, "new": true, "create": true, "change": true, "edit": true, "update": true,
	"delete": true, "remove": true, "view": true, "manage": true, "set": true, "make": true,
	"check": true, "list": true, "all": true,
}

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "i": true, "me": true, "my": true, "we": true, "our": true,
	"you": true, "your": true, "to": true, "of": true, "in": true, "on": true, "for": true, "at": true,
	"by": true, "with": true, "and": true, "or": true, "is": true, "are": true, "can": true, "do": true,
	"does": true, "how": true, "where": true, "what": true, "which": true, "page": true, "screen": true,
	"find": true, "go": true, "open": true, "there": true, "it": true, "this": true, "that": true,
	"rtylr": true, "portal": true, "please": true, "want": true, "need": true, "see": true, "get": true,
	"from": true, "about": true, "some": true, "someone": true, "who": true,
}

// terms splits text into lowercase stemmed words without stop words.
func terms(text string) map[string]bool {
	out := map[string]bool{}
	for _, word := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if !stopWords[word] {
			out[stem(word)] = true
		}
	}
	return out
}

// stem folds the common English endings so "members" finds "member" and
// "adding" finds "add".
func stem(word string) string {
	switch {
	case len(word) > 4 && strings.HasSuffix(word, "ies"):
		return strings.TrimSuffix(word, "ies") + "y"
	case len(word) > 3 && strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss"):
		word = strings.TrimSuffix(word, "s")
	}
	// After the plural, so "settings" and "setting" fold the same way.
	if len(word) > 5 && strings.HasSuffix(word, "ing") {
		word = strings.TrimSuffix(word, "ing")
	}
	return word
}
