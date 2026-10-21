package infrastructure

type cvProfileJSON struct {
	Skills          []string `json:"skills,omitempty"`
	Functions       []string `json:"functions,omitempty"`
	Industries      []string `json:"industries,omitempty"`
	RoleTitles      []string `json:"role_titles,omitempty"`
	Seniority       string   `json:"seniority,omitempty"`
	YearsExperience string   `json:"years_experience,omitempty"`
	Education       []string `json:"education,omitempty"`
	Languages       []string `json:"languages,omitempty"`
	Certifications  []string `json:"certifications,omitempty"`
	Locations       []string `json:"locations,omitempty"`
	Availability    string   `json:"availability,omitempty"`
	SalarySignals   []string `json:"salary_signals,omitempty"`
}
