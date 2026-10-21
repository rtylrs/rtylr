package infrastructure

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

// GetRecruitCareerByIDRow selects the same columns as
// GetRecruitCareerBySlugRow, so it shares the slug row's mapper (struct tags
// are ignored for the conversion).
func recruitCareerFromIDRow(row sqlc.GetRecruitCareerByIDRow) domain.RecruitCareer {
	return recruitCareerFromGetRow(sqlc.GetRecruitCareerBySlugRow(row))
}

func customizationFromBusinessRow(row sqlc.GetCustomizationByBusinessRow) domain.Customization {
	return domain.Customization{
		ID:                 row.ID,
		OwnerID:            row.OwnerID,
		BusinessID:         row.BusinessID,
		Slug:               row.Slug,
		BrandName:          row.BrandName,
		Tagline:            row.Tagline,
		AboutBlurb:         row.AboutBlurb,
		LogoURL:            row.LogoUrl,
		PrimaryColor:       row.PrimaryColor,
		AccentColor:        row.AccentColor,
		BgColor:            row.BgColor,
		TextColor:          row.TextColor,
		MutedTextColor:     row.MutedTextColor,
		FontPairing:        row.FontPairing,
		HeroImageURL:       row.HeroImageUrl,
		FaviconURL:         row.FaviconUrl,
		Template:           row.Template,
		HeaderHTML:         row.HeaderHtml,
		MetaTitle:          row.MetaTitle,
		MetaDescription:    row.MetaDescription,
		PrivacyURL:         row.PrivacyUrl,
		TermsURL:           row.TermsUrl,
		GAMeasurementID:    row.GaMeasurementID,
		MetaPixelID:        row.MetaPixelID,
		RequireConsent:     row.RequireConsent,
		NotificationEmail:  row.NotificationEmail,
		EmailNotifications: row.EmailNotifications,
		CreatedAt:          nullTimeValue(row.CreatedAt),
		UpdatedAt:          nullTimeValue(row.UpdatedAt),
	}
}

func customizationFromSlugRow(row sqlc.GetCustomizationBySlugRow) domain.Customization {
	return domain.Customization{
		ID:                 row.ID,
		OwnerID:            row.OwnerID,
		BusinessID:         row.BusinessID,
		Slug:               row.Slug,
		BrandName:          row.BrandName,
		Tagline:            row.Tagline,
		AboutBlurb:         row.AboutBlurb,
		LogoURL:            row.LogoUrl,
		PrimaryColor:       row.PrimaryColor,
		AccentColor:        row.AccentColor,
		BgColor:            row.BgColor,
		TextColor:          row.TextColor,
		MutedTextColor:     row.MutedTextColor,
		FontPairing:        row.FontPairing,
		HeroImageURL:       row.HeroImageUrl,
		FaviconURL:         row.FaviconUrl,
		Template:           row.Template,
		HeaderHTML:         row.HeaderHtml,
		MetaTitle:          row.MetaTitle,
		MetaDescription:    row.MetaDescription,
		PrivacyURL:         row.PrivacyUrl,
		TermsURL:           row.TermsUrl,
		GAMeasurementID:    row.GaMeasurementID,
		MetaPixelID:        row.MetaPixelID,
		RequireConsent:     row.RequireConsent,
		NotificationEmail:  row.NotificationEmail,
		EmailNotifications: row.EmailNotifications,
		CreatedAt:          nullTimeValue(row.CreatedAt),
		UpdatedAt:          nullTimeValue(row.UpdatedAt),
	}
}

func submissionsFromRows(rows []sqlc.ListCareerSubmissionsRow) []domain.Submission {
	out := make([]domain.Submission, 0, len(rows))
	for _, row := range rows {
		out = append(out, submissionFromRow(row))
	}
	return out
}

func submissionsFromStatusRows(rows []sqlc.ListCareerSubmissionsByStatusRow) []domain.Submission {
	out := make([]domain.Submission, 0, len(rows))
	for _, row := range rows {
		out = append(out, submissionFromStatusRow(row))
	}
	return out
}

// The paged rows select the same columns as their unpaged siblings, so they
// convert (struct tags are ignored) and share the same mapper.
func submissionsFromPageRows(rows []sqlc.ListCareerSubmissionsPageRow) []domain.Submission {
	out := make([]domain.Submission, 0, len(rows))
	for _, row := range rows {
		out = append(out, submissionFromRow(sqlc.ListCareerSubmissionsRow(row)))
	}
	return out
}

func submissionsFromStatusPageRows(rows []sqlc.ListCareerSubmissionsByStatusPageRow) []domain.Submission {
	out := make([]domain.Submission, 0, len(rows))
	for _, row := range rows {
		out = append(out, submissionFromStatusRow(sqlc.ListCareerSubmissionsByStatusRow(row)))
	}
	return out
}

func submissionFromRow(row sqlc.ListCareerSubmissionsRow) domain.Submission {
	return domain.Submission{
		ID:            row.ID,
		CareerID:      row.CareerID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		CandidateName: row.CandidateName,
		Email:         row.Email,
		Phone:         row.Phone,
		ResumeURL:     row.ResumeUrl,
		CoverLetter:   row.CoverLetter,
		Status:        row.Status,
		AnswersJSON:   interfaceToString(row.AnswersJson),
		CreatedAt:     nullTimeValue(row.CreatedAt),
		UpdatedAt:     nullTimeValue(row.UpdatedAt),
		CareerTitle:   row.CareerTitle,
	}
}

func submissionFromStatusRow(row sqlc.ListCareerSubmissionsByStatusRow) domain.Submission {
	return domain.Submission{
		ID:            row.ID,
		CareerID:      row.CareerID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		CandidateName: row.CandidateName,
		Email:         row.Email,
		Phone:         row.Phone,
		ResumeURL:     row.ResumeUrl,
		CoverLetter:   row.CoverLetter,
		Status:        row.Status,
		AnswersJSON:   interfaceToString(row.AnswersJson),
		CreatedAt:     nullTimeValue(row.CreatedAt),
		UpdatedAt:     nullTimeValue(row.UpdatedAt),
		CareerTitle:   row.CareerTitle,
	}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullDecimal(amount money.Amount) sql.NullString {
	if amount.IsZero() {
		return sql.NullString{}
	}
	return sql.NullString{String: amount.String(), Valid: true}
}

func parseSalary(v string) money.Amount {
	if v == "" {
		return 0
	}
	amount, err := money.FromString(v)
	if err != nil {
		return 0
	}
	return amount
}
