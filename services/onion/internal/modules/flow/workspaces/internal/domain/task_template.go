package domain

import (
	"errors"
	"time"
)

// TaskTemplate is a project's description template for one task type (T-380).
type TaskTemplate struct {
	ID        int64
	ProjectID int64
	TaskType  string
	Body      string
	UpdatedAt time.Time
}

// MaxTaskTemplateLength caps a template body, in bytes.
const MaxTaskTemplateLength = 20000

// ErrTaskTemplateTooLong rejects a body over MaxTaskTemplateLength.
var ErrTaskTemplateTooLong = errors.New("template is too long")
