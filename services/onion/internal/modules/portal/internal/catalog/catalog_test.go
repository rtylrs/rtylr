package catalog

import "testing"

func TestSearchFindsThePageForWhatPeopleAsk(t *testing.T) {
	cases := map[string]string{
		"add team members":            "/team",
		"where can I invite someone":  "/team",
		"add a customer":              "/crm/customers",
		"create an invoice":           "/finance/invoices",
		"change my password":          "/security",
		"create an api key":           "/api-keys",
		"my tasks":                    "/flow/my-issues",
		"deals pipeline":              "/crm/sales",
		"mark notifications as read":  "/notifications",
		"set up a routine with caret": "/caret",
	}
	for query, want := range cases {
		got := Search(query, 3)
		found := false
		for _, page := range got {
			found = found || page.Path == want
		}
		if !found {
			paths := []string{}
			for _, page := range got {
				paths = append(paths, page.Path)
			}
			t.Errorf("%q: want %s in the top 3, got %v", query, want, paths)
		}
	}
	if got := Search("where is it", 3); len(got) != 0 {
		t.Errorf("stop words alone matched %d pages", len(got))
	}
}

func TestStemFoldsCommonEndings(t *testing.T) {
	for word, want := range map[string]string{"members": "member", "adding": "add", "categories": "category", "address": "address", "settings": "sett", "setting": "sett"} {
		if got := stem(word); got != want {
			t.Errorf("stem(%q) = %q, want %q", word, got, want)
		}
	}
}
