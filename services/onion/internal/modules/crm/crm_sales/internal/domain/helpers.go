package domain

import "github.com/voxire/onion/internal/platform/money"

// AllStages returns the ordered list of lead stages from earliest to latest.
func AllStages() []LeadStage {
	return []LeadStage{
		LeadStageCold,
		LeadStageContacted,
		LeadStageDemoBooked,
		LeadStageDemoDone,
		LeadStageProposalSent,
		LeadStageContracted,
		LeadStageLost,
	}
}

// IsOpen returns true if the lead is still in an active (non-terminal) stage.
func (l *Lead) IsOpen() bool {
	return l.Stage != LeadStageContracted && l.Stage != LeadStageLost
}

// IsWon returns true if the lead ended in a contract.
func (l *Lead) IsWon() bool {
	return l.Stage == LeadStageContracted
}

// IsLost returns true if the lead was marked as lost.
func (l *Lead) IsLost() bool {
	return l.Stage == LeadStageLost
}

// HasMRR returns true if an MRR potential was set on the lead.
func (l *Lead) HasMRR() bool {
	return l.MRRPotential != nil && *l.MRRPotential > 0
}

// TotalPipelineMRR sums the MRR across a set of pipeline summaries.
func TotalPipelineMRR(summaries []PipelineSummary) money.Amount {
	var total money.Amount
	for _, s := range summaries {
		total += s.TotalMRR
	}
	return total
}

// TotalPipelineCount sums the lead count across all pipeline stages.
func TotalPipelineCount(summaries []PipelineSummary) int {
	total := 0
	for _, s := range summaries {
		total += s.Count
	}
	return total
}

// HasNotes returns true when the lead has notes attached.
func (l *Lead) HasNotes() bool {
	return l.Notes != ""
}

// HasAssignee returns true when the lead has been assigned to a team member.
func (l *Lead) HasAssignee() bool {
	return l.AssignedTo != nil
}

// HasEmail returns true when the contact has an email address.
func (c *Contact) HasEmail() bool {
	return c.Email != ""
}

// HasPhone returns true when the contact has a phone number.
func (c *Contact) HasPhone() bool {
	return c.Phone != ""
}

// HasLinkedIn returns true when the contact has a LinkedIn profile set.
func (c *Contact) HasLinkedIn() bool {
	return c.LinkedIn != ""
}

// IsPrimaryContact returns true when this contact is the primary contact for a lead.
func (c *Contact) IsPrimaryContact() bool {
	return c.IsPrimary
}

// HasNextFollowUp returns true when the touchpoint has a scheduled follow-up.
func (tp *Touchpoint) HasNextFollowUp() bool {
	return tp.NextFollowUp != nil && *tp.NextFollowUp != ""
}

// HasContact returns true when the touchpoint is linked to a specific contact.
func (tp *Touchpoint) HasContact() bool {
	return tp.ContactID != nil
}

// ValidateCreateLeadRequest checks required fields.
// Returns a non-empty string if the request is invalid.
func ValidateCreateLeadRequest(req CreateLeadRequest) string {
	if req.CompanyName == "" {
		return "company name is required"
	}
	return ""
}

// ValidateCreateContractRequest checks required fields for creating a contract.
func ValidateCreateContractRequest(req CreateContractRequest) string {
	if req.LeadID <= 0 {
		return "lead_id is required"
	}
	if req.Plan == "" {
		return "plan is required"
	}
	return ""
}

// HasMRRValue returns true when the contract has a non-zero MRR value.
func (c *Contract) HasMRRValue() bool {
	return c.MRR > 0
}

// HasSignedBy returns true when the contract records who signed it.
func (c *Contract) HasSignedBy() bool {
	return c.SignedByName != ""
}
