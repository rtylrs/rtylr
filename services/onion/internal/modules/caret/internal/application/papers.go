package application

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

var papersEndpoint = "https://api.openalex.org/works"

func init() {
	registerGeneralTool("knowledge.papers", (*AssistantToolExecutor).searchPapers)
}

// searchPapers finds scholarly work on a topic in OpenAlex, for questions
// that deserve evidence rather than blog posts.
func (e *AssistantToolExecutor) searchPapers(ctx context.Context, input map[string]any) (map[string]any, error) {
	query := inputText(input, "query", "topic", "q")
	if query == "" || len([]rune(query)) > 200 {
		return nil, errors.New("give a research topic")
	}
	params := url.Values{
		"search":   {query},
		"per-page": {"6"},
		"select":   {"display_name,publication_year,cited_by_count,doi,primary_location,authorships,abstract_inverted_index"},
	}
	if year, ok := inputNumber(input, "since", "from_year"); ok && year > 1900 {
		params.Set("filter", "from_publication_date:"+strconv.Itoa(int(year))+"-01-01")
	}
	if inputBool(input, "most_cited") {
		params.Set("sort", "cited_by_count:desc")
	}
	var response struct {
		Results []struct {
			Title    string `json:"display_name"`
			Year     int    `json:"publication_year"`
			Cited    int    `json:"cited_by_count"`
			DOI      string `json:"doi"`
			Location struct {
				LandingPage string `json:"landing_page_url"`
				Source      *struct {
					Name string `json:"display_name"`
				} `json:"source"`
			} `json:"primary_location"`
			Authorships []struct {
				Author struct {
					Name string `json:"display_name"`
				} `json:"author"`
			} `json:"authorships"`
			Abstract map[string][]int `json:"abstract_inverted_index"`
		} `json:"results"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), papersEndpoint+"?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, errors.New("no papers match that topic")
	}
	papers := make([]map[string]any, 0, len(response.Results))
	sources := []map[string]string{}
	for _, work := range response.Results {
		authors := []string{}
		for i, authorship := range work.Authorships {
			if i == 3 {
				authors = append(authors, "et al.")
				break
			}
			authors = append(authors, authorship.Author.Name)
		}
		link := work.DOI
		if link == "" {
			link = work.Location.LandingPage
		}
		paper := map[string]any{"title": work.Title, "year": work.Year, "cited_by": work.Cited, "authors": strings.Join(authors, ", "), "url": link,
			"abstract": truncateSearchText(rebuildAbstract(work.Abstract), 900)}
		if work.Location.Source != nil {
			paper["venue"] = work.Location.Source.Name
		}
		papers = append(papers, paper)
		if strings.HasPrefix(link, "https://") {
			sources = append(sources, map[string]string{"title": work.Title, "url": link})
		}
	}
	return map[string]any{"query": query, "papers": papers, "sources": sources}, nil
}

// rebuildAbstract turns OpenAlex's word-position index back into text.
func rebuildAbstract(index map[string][]int) string {
	type placed struct {
		at   int
		word string
	}
	words := []placed{}
	for word, positions := range index {
		for _, at := range positions {
			words = append(words, placed{at, word})
		}
	}
	sort.Slice(words, func(i, j int) bool { return words[i].at < words[j].at })
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = w.word
	}
	return strings.Join(out, " ")
}
