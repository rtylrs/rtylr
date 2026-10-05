package domain

// HasDescription returns true if the segment has a non-empty description.
func (s *Segment) HasDescription() bool {
	return s.Description != ""
}

// HasCustomers returns true if the segment has at least one member.
func (s *Segment) HasCustomers() bool {
	return s.CustomerCount > 0
}

// HasFilters returns true if the segment has non-empty filter rules.
func (s *Segment) HasFilters() bool {
	return len(s.Filters) > 0 && string(s.Filters) != "{}"
}

// ValidateSegmentInput checks required fields.
// Returns a non-empty string if validation fails.
func ValidateSegmentInput(input SegmentInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.Name == "" {
		return "segment name is required"
	}
	return ""
}

// ValidateSegmentUpdateInput checks required fields for an update.
func ValidateSegmentUpdateInput(input SegmentUpdateInput) string {
	if input.ID <= 0 {
		return "segment id is required"
	}
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	if input.Name == "" {
		return "segment name is required"
	}
	return ""
}

// IsPopulated returns true when the segment has been calculated and has at least one customer.
func (s *Segment) IsPopulated() bool {
	return s.CustomerCount > 0
}

// IsEmpty returns true when the segment has been calculated and has no customers.
func (s *Segment) IsEmpty() bool {
	return s.CustomerCount == 0
}
