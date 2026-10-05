package domain

import "testing"

func TestContactHasEmail_WhenSet(t *testing.T) {
	c := Contact{Email: "alice@example.com"}
	if !c.HasEmail() {
		t.Error("expected HasEmail=true when email is set")
	}
}

func TestContactHasEmail_WhenEmpty(t *testing.T) {
	c := Contact{Email: ""}
	if c.HasEmail() {
		t.Error("expected HasEmail=false when email is empty")
	}
}

func TestContactHasPhone_WhenSet(t *testing.T) {
	c := Contact{Phone: "+44207123456"}
	if !c.HasPhone() {
		t.Error("expected HasPhone=true when phone is set")
	}
}

func TestContactHasPhone_WhenEmpty(t *testing.T) {
	c := Contact{Phone: ""}
	if c.HasPhone() {
		t.Error("expected HasPhone=false when phone is empty")
	}
}

func TestContactHasLinkedIn_WhenSet(t *testing.T) {
	c := Contact{LinkedIn: "https://linkedin.com/in/alice"}
	if !c.HasLinkedIn() {
		t.Error("expected HasLinkedIn=true when linkedin is set")
	}
}

func TestContactHasLinkedIn_WhenEmpty(t *testing.T) {
	c := Contact{LinkedIn: ""}
	if c.HasLinkedIn() {
		t.Error("expected HasLinkedIn=false when linkedin is empty")
	}
}

func TestContactIsPrimaryContact_WhenTrue(t *testing.T) {
	c := Contact{IsPrimary: true}
	if !c.IsPrimaryContact() {
		t.Error("expected IsPrimaryContact=true")
	}
}

func TestContactIsPrimaryContact_WhenFalse(t *testing.T) {
	c := Contact{IsPrimary: false}
	if c.IsPrimaryContact() {
		t.Error("expected IsPrimaryContact=false")
	}
}

func TestTouchpointHasNextFollowUp_WhenSet(t *testing.T) {
	fu := "2026-06-15"
	tp := Touchpoint{NextFollowUp: &fu}
	if !tp.HasNextFollowUp() {
		t.Error("expected HasNextFollowUp=true when follow-up date is set")
	}
}

func TestTouchpointHasNextFollowUp_WhenNil(t *testing.T) {
	tp := Touchpoint{NextFollowUp: nil}
	if tp.HasNextFollowUp() {
		t.Error("expected HasNextFollowUp=false when follow-up is nil")
	}
}

func TestTouchpointHasContact_WhenSet(t *testing.T) {
	id := int64(10)
	tp := Touchpoint{ContactID: &id}
	if !tp.HasContact() {
		t.Error("expected HasContact=true when contact_id is set")
	}
}

func TestTouchpointHasContact_WhenNil(t *testing.T) {
	tp := Touchpoint{ContactID: nil}
	if tp.HasContact() {
		t.Error("expected HasContact=false when contact_id is nil")
	}
}
