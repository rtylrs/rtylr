package domain

// HasFromEmail returns true if a sender email address is configured.
func (s *Settings) HasFromEmail() bool {
	return s.FromEmail != ""
}

// HasFromName returns true if a sender display name is configured.
func (s *Settings) HasFromName() bool {
	return s.FromName != ""
}

// HasSmsSenderID returns true if an SMS sender ID is configured.
func (s *Settings) HasSmsSenderID() bool {
	return s.SmsSenderID != ""
}

// HasTimezone returns true if a timezone has been set.
func (s *Settings) HasTimezone() bool {
	return s.Timezone != ""
}

// ConsentRequired returns true when new contacts require explicit opt-in by default.
func (s *Settings) ConsentRequired() bool {
	return !s.DefaultConsent
}

// ValidateSettingsInput checks required fields for upserting settings.
func ValidateSettingsInput(input SettingsInput) string {
	if input.OwnerID <= 0 {
		return "owner_id is required"
	}
	return ""
}

// IsFullyConfigured returns true when all critical sending fields are set.
func (s *Settings) IsFullyConfigured() bool {
	return s.FromEmail != "" && s.FromName != "" && s.Timezone != ""
}

// HasSmsCapability returns true when the settings have an SMS sender ID configured.
func (s *Settings) HasSmsCapability() bool {
	return s.SmsSenderID != ""
}
