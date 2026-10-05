package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

func TestRecruitResponseExcludesCredentialsAndPreservesNestedProfile(t *testing.T) {
	admin := &domain.Admin{ID: 12, Name: "Recruiter", Email: "recruiter@example.test", PasswordHash: "private-credential"}
	candidate := domain.Candidate{
		ID: 34, Name: "Candidate",
		CVProfile:   domain.CVProfile{RoleTitles: []string{"Engineer"}, YearsExperience: "5"},
		MatchSignal: domain.MatchSignal{Score: 80, MatchedTerms: []string{"Go"}},
	}
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]any{
		"admin":     adminViewPointer(admin),
		"candidate": candidateViewOf(candidate),
	})
	if strings.Contains(w.Body.String(), admin.PasswordHash) {
		t.Fatal("recruiter credential leaked into response")
	}
	var body struct {
		Admin struct {
			Email string `json:"email"`
		} `json:"admin"`
		Candidate struct {
			Profile struct {
				Roles      []string `json:"role_titles"`
				Experience string   `json:"years_experience"`
			} `json:"cv_profile"`
			Match struct {
				Terms []string `json:"matched_terms"`
			} `json:"match_signal"`
		} `json:"candidate"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Admin.Email != admin.Email || len(body.Candidate.Profile.Roles) != 1 || body.Candidate.Profile.Roles[0] != "Engineer" || body.Candidate.Profile.Experience != "5" || len(body.Candidate.Match.Terms) != 1 || body.Candidate.Match.Terms[0] != "Go" {
		t.Fatalf("response lost its wire fields: %s", w.Body.String())
	}
}
