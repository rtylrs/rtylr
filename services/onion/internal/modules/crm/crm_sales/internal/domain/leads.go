package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/voxire/onion/internal/platform/money"
)

// LeadStatus is where prospecting work on a lead stands.
type LeadStatus string

const (
	LeadNew          LeadStatus = "new"
	LeadAttempting   LeadStatus = "attempting"
	LeadConnected    LeadStatus = "connected"
	LeadNurture      LeadStatus = "nurture"
	LeadQualified    LeadStatus = "qualified"
	LeadDisqualified LeadStatus = "disqualified"
)

// LeadStatuses lists every status in working order.
func LeadStatuses() []LeadStatus {
	return []LeadStatus{LeadNew, LeadAttempting, LeadConnected, LeadNurture, LeadQualified, LeadDisqualified}
}

func (s LeadStatus) IsValid() bool {
	for _, status := range LeadStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// IsWorking reports whether the lead is still being prospected.
func (s LeadStatus) IsWorking() bool {
	return s == LeadNew || s == LeadAttempting || s == LeadConnected || s == LeadNurture
}

// Lead is a prospect to qualify: a person and the company they work for.
type Lead struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	FirstName        string
	LastName         string
	JobTitle         string
	Email            string
	Phone            string
	CompanyName      string
	Website          string
	LinkedIn         string
	Industry         string
	Size             string
	City             string
	Country          string
	Source           string
	Status           LeadStatus
	DisqualifyReason string
	AssignedTo       *int64
	Value            *money.Amount
	Billing          Billing
	Notes            string
	AttemptCount     int
	LastActivityAt   *time.Time
	// NextFollowUp is the earliest open task's due date (YYYY-MM-DD), or "".
	NextFollowUp  string
	CustomerID    *int64
	CompanyID     *int64
	DealID        *int64
	ConvertedAt   *time.Time
	ImportBatchID *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PersonName is the lead's full name, or "" when only a company is known.
func (l Lead) PersonName() string {
	return strings.TrimSpace(l.FirstName + " " + l.LastName)
}

// DisplayName names the lead in lists and deal titles.
func (l Lead) DisplayName() string {
	if l.CompanyName != "" {
		return l.CompanyName
	}
	return l.PersonName()
}

// LeadInput is a lead's editable fields.
type LeadInput struct {
	FirstName   string
	LastName    string
	JobTitle    string
	Email       string
	Phone       string
	CompanyName string
	Website     string
	LinkedIn    string
	Industry    string
	Size        string
	City        string
	Country     string
	Source      string
	AssignedTo  *int64
	Value       *money.Amount
	Billing     Billing
	Notes       string
}

// Normalize trims and bounds every field to its column and checks the lead
// names someone: a company or a person.
func (in *LeadInput) Normalize() error {
	in.FirstName = trimTo(in.FirstName, 100)
	in.LastName = trimTo(in.LastName, 100)
	in.JobTitle = trimTo(in.JobTitle, 150)
	in.Email = strings.ToLower(trimTo(in.Email, 255))
	in.Phone = trimTo(in.Phone, 50)
	in.CompanyName = trimTo(in.CompanyName, 255)
	in.Website = trimTo(in.Website, 255)
	in.LinkedIn = trimTo(in.LinkedIn, 255)
	in.Industry = trimTo(in.Industry, 100)
	in.Size = trimTo(in.Size, 50)
	in.City = trimTo(in.City, 100)
	in.Country = trimTo(in.Country, 100)
	in.Source = trimTo(in.Source, 100)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.CompanyName == "" && in.FirstName == "" && in.LastName == "" {
		return invalid("a lead needs a company or a person's name")
	}
	if in.Email != "" && !looksLikeEmail(in.Email) {
		return invalid("%q is not an email address", in.Email)
	}
	if in.Phone != "" && PhoneDigits(in.Phone) == "" {
		return invalid("%q is not a phone number", in.Phone)
	}
	if in.Billing == "" {
		in.Billing = BillingOneTime
	}
	if !in.Billing.IsValid() {
		return invalid("billing must be one_time, monthly or yearly")
	}
	if in.Value != nil && *in.Value < 0 {
		return invalid("value cannot be negative")
	}
	if in.AssignedTo != nil && *in.AssignedTo <= 0 {
		in.AssignedTo = nil
	}
	return nil
}

// PhoneDigits is a phone number reduced to its digits, the form duplicates
// are matched on. Formatting and a leading + are dropped.
func PhoneDigits(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if len(digits) > 32 {
		digits = digits[:32]
	}
	return digits
}

// LeadSort orders a lead list.
type LeadSort string

const (
	LeadSortNewest   LeadSort = "newest"
	LeadSortFollowUp LeadSort = "follow_up"
	LeadSortActivity LeadSort = "recent_activity"
	LeadSortCompany  LeadSort = "company"
)

func (s LeadSort) IsValid() bool {
	return s == LeadSortNewest || s == LeadSortFollowUp || s == LeadSortActivity || s == LeadSortCompany
}

// FollowUpWindow narrows a list to work due in a window around today.
type FollowUpWindow string

const (
	FollowUpAny      FollowUpWindow = ""
	FollowUpOverdue  FollowUpWindow = "overdue"
	FollowUpToday    FollowUpWindow = "today"
	FollowUpDue      FollowUpWindow = "due"
	FollowUpUpcoming FollowUpWindow = "upcoming"
)

func (w FollowUpWindow) IsValid() bool {
	switch w {
	case FollowUpAny, FollowUpOverdue, FollowUpToday, FollowUpDue, FollowUpUpcoming:
		return true
	}
	return false
}

// Range is the window's inclusive date bounds relative to today; nil means
// unbounded.
func (w FollowUpWindow) Range(today time.Time) (from, to *time.Time) {
	day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := day.AddDate(0, 0, -1)
	tomorrow := day.AddDate(0, 0, 1)
	switch w {
	case FollowUpOverdue:
		return nil, &yesterday
	case FollowUpToday:
		return &day, &day
	case FollowUpDue:
		return nil, &day
	case FollowUpUpcoming:
		return &tomorrow, nil
	}
	return nil, nil
}

// LeadFilter selects leads. AssignedTo 0 means unassigned; empty Statuses
// means every status.
type LeadFilter struct {
	Statuses      []LeadStatus
	AssignedTo    *int64
	Source        string
	ImportBatchID *int64
	FollowUp      FollowUpWindow
	Search        string
	Sort          LeadSort
	Cursor        string
	Limit         int
}

// LeadPage is one page of a lead list. NextCursor is "" on the last page.
type LeadPage struct {
	Leads      []Lead
	NextCursor string
	Total      int64
}

// LeadStats summarises prospecting for the leads page header.
type LeadStats struct {
	ByStatus           map[LeadStatus]int64
	FollowUpsOverdue   int64
	FollowUpsToday     int64
	CallsToday         int64
	ConnectedToday     int64
	ConvertedThisMonth int64
}

// LeadChanges is a full edit of a lead, status included.
type LeadChanges struct {
	LeadInput
	Status           LeadStatus
	DisqualifyReason string
}

// Normalize checks the fields and the status change. Qualifying happens
// through conversion only.
func (c *LeadChanges) Normalize(current LeadStatus) error {
	if err := c.LeadInput.Normalize(); err != nil {
		return err
	}
	if c.Status == "" {
		c.Status = current
	}
	if !c.Status.IsValid() {
		return invalid("unknown lead status %q", c.Status)
	}
	if c.Status == LeadQualified && current != LeadQualified {
		return invalid("convert the lead to qualify it")
	}
	if current == LeadQualified && c.Status != LeadQualified {
		return ErrLeadConverted
	}
	c.DisqualifyReason = trimTo(c.DisqualifyReason, 255)
	if c.Status == LeadDisqualified && c.DisqualifyReason == "" {
		return invalid("a disqualified lead needs a reason")
	}
	if c.Status != LeadDisqualified {
		c.DisqualifyReason = ""
	}
	return nil
}

// BulkLeadUpdate changes many leads at once: who works them or their status.
type BulkLeadUpdate struct {
	IDs              []int64
	AssignTo         *int64
	Unassign         bool
	Status           LeadStatus
	DisqualifyReason string
}

func (u *BulkLeadUpdate) Normalize() error {
	if len(u.IDs) == 0 {
		return invalid("select at least one lead")
	}
	if len(u.IDs) > 1000 {
		return invalid("update at most 1000 leads at a time")
	}
	changes := 0
	if u.AssignTo != nil || u.Unassign {
		changes++
	}
	if u.Status != "" {
		changes++
		if !u.Status.IsValid() || u.Status == LeadQualified {
			return invalid("leads can be set to new, attempting, connected, nurture or disqualified")
		}
		u.DisqualifyReason = trimTo(u.DisqualifyReason, 255)
		if u.Status == LeadDisqualified && u.DisqualifyReason == "" {
			return invalid("disqualifying needs a reason")
		}
	}
	if changes != 1 {
		return invalid("choose either an assignee or a status")
	}
	return nil
}

// ConvertLeadInput qualifies a lead. With a deal, the lead's work continues in
// the pipeline; without, only the person and company are created.
type ConvertLeadInput struct {
	CreateDeal bool
	Deal       DealInput
}

func trimTo(value string, maxRunes int) string {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	return strings.TrimSpace(string([]rune(value)[:maxRunes]))
}

func looksLikeEmail(value string) bool {
	at := strings.LastIndex(value, "@")
	return at > 0 && at < len(value)-1 && !strings.ContainsAny(value, " \t\r\n,;")
}

// IsDate reports whether value is a calendar date in YYYY-MM-DD form.
func IsDate(value string) bool {
	_, err := time.Parse(time.DateOnly, value)
	return err == nil
}
