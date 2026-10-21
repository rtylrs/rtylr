package domain

import "time"

// ActivityType is the kind of touch logged against a lead or deal.
type ActivityType string

const (
	ActivityCall    ActivityType = "call"
	ActivityEmail   ActivityType = "email"
	ActivityMeeting ActivityType = "meeting"
	ActivitySMS     ActivityType = "sms"
	ActivityNote    ActivityType = "note"
)

func (t ActivityType) IsValid() bool {
	switch t {
	case ActivityCall, ActivityEmail, ActivityMeeting, ActivitySMS, ActivityNote:
		return true
	}
	return false
}

// CallOutcome is how a call ended: the dispositions a calling list needs.
type CallOutcome string

const (
	CallConnected     CallOutcome = "connected"
	CallNoAnswer      CallOutcome = "no_answer"
	CallVoicemail     CallOutcome = "voicemail"
	CallBusy          CallOutcome = "busy"
	CallWrongNumber   CallOutcome = "wrong_number"
	CallCallback      CallOutcome = "callback"
	CallMeetingBooked CallOutcome = "meeting_booked"
	CallNotInterested CallOutcome = "not_interested"
)

func (o CallOutcome) IsValid() bool {
	switch o {
	case CallConnected, CallNoAnswer, CallVoicemail, CallBusy, CallWrongNumber,
		CallCallback, CallMeetingBooked, CallNotInterested:
		return true
	}
	return false
}

// Reached reports whether the call got through to the person.
func (o CallOutcome) Reached() bool {
	return o == CallConnected || o == CallCallback || o == CallMeetingBooked || o == CallNotInterested
}

// NextLeadStatus is the lead's status after an activity, so the list
// reflects the work without bookkeeping. Reaching the person makes the lead
// connected; trying without reaching them makes a new lead attempting.
// Qualified and disqualified leads are only changed by people.
func NextLeadStatus(current LeadStatus, activity ActivityType, outcome CallOutcome) LeadStatus {
	if !current.IsWorking() {
		return current
	}
	switch activity {
	case ActivityCall:
		if outcome.Reached() {
			return LeadConnected
		}
		if current == LeadNew && outcome != CallWrongNumber {
			return LeadAttempting
		}
	case ActivityMeeting:
		return LeadConnected
	case ActivityEmail, ActivitySMS:
		if current == LeadNew {
			return LeadAttempting
		}
	}
	return current
}

// Activity is one logged touch.
type Activity struct {
	ID              int64
	BusinessID      int64
	LeadID          *int64
	DealID          *int64
	Type            ActivityType
	Outcome         CallOutcome
	Notes           string
	DurationSeconds *int
	CreatedBy       int64
	OccurredAt      time.Time
	CreatedAt       time.Time
}

// ActivityInput logs a touch on a lead or a deal, optionally scheduling the
// next one as a task.
type ActivityInput struct {
	LeadID          *int64
	DealID          *int64
	Type            ActivityType
	Outcome         CallOutcome
	Notes           string
	DurationSeconds *int
	OccurredAt      *time.Time
	FollowUp        *TaskInput
}

func (in *ActivityInput) Normalize(now time.Time) error {
	if (in.LeadID == nil) == (in.DealID == nil) {
		return invalid("log the activity on either a lead or a deal")
	}
	if !in.Type.IsValid() {
		return invalid("unknown activity type %q", in.Type)
	}
	if in.Type == ActivityCall {
		if !in.Outcome.IsValid() {
			return invalid("a call needs an outcome")
		}
	} else {
		in.Outcome = ""
		in.DurationSeconds = nil
	}
	if in.DurationSeconds != nil && (*in.DurationSeconds < 0 || *in.DurationSeconds > 24*60*60) {
		return invalid("call duration must be between 0 and 24 hours")
	}
	in.Notes = trimTo(in.Notes, 10000)
	if in.Type == ActivityNote && in.Notes == "" {
		return invalid("a note needs text")
	}
	if in.OccurredAt == nil || in.OccurredAt.After(now.Add(5*time.Minute)) {
		at := now
		in.OccurredAt = &at
	}
	if in.FollowUp != nil {
		in.FollowUp.LeadID = in.LeadID
		in.FollowUp.DealID = in.DealID
		if err := in.FollowUp.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

// ActivityPage is one page of a timeline, newest first.
type ActivityPage struct {
	Activities []Activity
	NextCursor string
}
