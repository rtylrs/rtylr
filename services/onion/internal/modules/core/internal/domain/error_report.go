package domain

type ErrorReport struct {
	Source         string
	Level          string
	Message        string
	Stack          string
	ComponentStack string
	URL            string
	Args           []interface{}
	Meta           map[string]interface{}
}

type ErrorReportRequest struct {
	Source         string
	Level          string
	Message        string
	Stack          string
	ComponentStack string
	URL            string
	Args           []interface{}
	Meta           map[string]interface{}
}
