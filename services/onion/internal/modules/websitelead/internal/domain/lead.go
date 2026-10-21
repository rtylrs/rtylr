package domain

type Lead struct {
	ID      int64
	Name    string
	Email   string
	Phone   string
	Company string
	Service string
	Project string
	Budget  string
	Source  string
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
