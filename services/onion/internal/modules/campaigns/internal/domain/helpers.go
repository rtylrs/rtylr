package domain

const (
	StatusDraft     = "draft"
	StatusScheduled = "scheduled"
	StatusSent      = "sent"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"

	TypeEmail = "email"
	TypeSMS   = "sms"
	TypePush  = "push"
)

// IsSent returns true if the campaign has been sent.
func (c *Campaign) IsSent() bool {
	return c.Status == StatusSent
}

// IsDraft returns true if the campaign is still in draft.
func (c *Campaign) IsDraft() bool {
	return c.Status == StatusDraft
}

// IsScheduled returns true if the campaign is scheduled to be sent.
func (c *Campaign) IsScheduled() bool {
	return c.Status == StatusScheduled
}

// OpenRate returns the ratio of opened messages to sent messages.
// Returns 0 if no messages were sent.
func (c *Campaign) OpenRate() float64 {
	if c.SentCount == 0 {
		return 0
	}
	return float64(c.OpenCount) / float64(c.SentCount)
}

// HasAudience returns true if the campaign targets a specific segment.
func (c *Campaign) HasAudience() bool {
	return c.AudienceSegmentID != nil
}

// IsCancelled returns true if the campaign has been cancelled.
func (c *Campaign) IsCancelled() bool {
	return c.Status == StatusCancelled
}

// HasBody returns true if the campaign has content in the body field.
func (c *Campaign) HasBody() bool {
	return c.Body != ""
}

// HasSubject returns true if the campaign has a subject line.
func (c *Campaign) HasSubject() bool {
	return c.Subject != ""
}
