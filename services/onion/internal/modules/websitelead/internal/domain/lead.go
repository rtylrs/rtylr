package domain

type Lead struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
	Service string `json:"service"`
	Project string `json:"project"`
	Budget  string `json:"budget"`
	Source  string `json:"source"`
}

type LeadInput struct {
	Name    string
	Email   string
	Phone   string
	Company string
	Service string
	Project string
	Budget  string
	Source  string
}
