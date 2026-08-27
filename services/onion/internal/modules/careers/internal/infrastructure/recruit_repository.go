package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

// ErrCareerNotFound / ErrCustomizationNotFound are sentinel errors the
// service layer translates into HTTP 404s.
var (
	ErrCareerNotFound        = errors.New("career not found")
	ErrCustomizationNotFound = errors.New("customization not found")
)

func (r *Repository) ListCareers(ctx context.Context, businessID int64, publishedOnly bool) ([]domain.RecruitCareer, error) {
	var rows []sqlc.ListRecruitCareersRow
	var err error
	if publishedOnly {
		publishedRows, listErr := r.queries.ListRecruitCareersPublished(ctx, businessID)
		if listErr != nil {
			return nil, listErr
		}
		rows = publishedToListRows(publishedRows)
	} else {
		rows, err = r.queries.ListRecruitCareers(ctx, businessID)
	}
	if err != nil {
		return nil, err
	}

	out := make([]domain.RecruitCareer, 0, len(rows))
	for _, row := range rows {
		out = append(out, recruitCareerFromRow(row))
	}
	return out, nil
}

func publishedToListRows(rows []sqlc.ListRecruitCareersPublishedRow) []sqlc.ListRecruitCareersRow {
	out := make([]sqlc.ListRecruitCareersRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, sqlc.ListRecruitCareersRow(row))
	}
	return out
}

func (r *Repository) GetCareerBySlug(ctx context.Context, businessID int64, slug string) (domain.RecruitCareer, error) {
	row, err := r.queries.GetRecruitCareerBySlug(ctx, sqlc.GetRecruitCareerBySlugParams{
		BusinessID: businessID,
		Slug:       slug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RecruitCareer{}, ErrCareerNotFound
		}
		return domain.RecruitCareer{}, err
	}
	return recruitCareerFromGetRow(row), nil
}

func (r *Repository) CreateCareer(ctx context.Context, in domain.RecruitCareerInput) (int64, error) {
	result, err := r.queries.CreateRecruitCareer(ctx, sqlc.CreateRecruitCareerParams{
		OwnerID:        in.OwnerID,
		BusinessID:     in.BusinessID,
		Slug:           in.Slug,
		Title:          in.Title,
		Department:     nullStr(in.Department),
		Location:       nullStr(in.Location),
		EmploymentType: nullStr(in.EmploymentType),
		WorkMode:       nullStr(in.WorkMode),
		Seniority:      nullStr(in.Seniority),
		Details:        in.Details,
		Requirements:   nullStr(in.Requirements),
		SalaryMin:      nullDecimal(in.SalaryMin),
		SalaryMax:      nullDecimal(in.SalaryMax),
		SalaryCurrency: nullStr(in.SalaryCurrency),
		IsPublished:    in.IsPublished,
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateCareer(ctx context.Context, in domain.RecruitCareerInput) error {
	businessID := in.BusinessID
	if businessID == 0 {
		businessID = in.OwnerID
	}

	return r.queries.UpdateRecruitCareer(ctx, sqlc.UpdateRecruitCareerParams{
		Slug:           in.Slug,
		Title:          in.Title,
		Department:     nullStr(in.Department),
		Location:       nullStr(in.Location),
		EmploymentType: nullStr(in.EmploymentType),
		WorkMode:       nullStr(in.WorkMode),
		Seniority:      nullStr(in.Seniority),
		Details:        in.Details,
		Requirements:   nullStr(in.Requirements),
		SalaryMin:      nullDecimal(in.SalaryMin),
		SalaryMax:      nullDecimal(in.SalaryMax),
		SalaryCurrency: nullStr(in.SalaryCurrency),
		IsPublished:    in.IsPublished,
		ID:             in.ID,
		BusinessID:     businessID,
	})
}

func (r *Repository) DeleteCareer(ctx context.Context, businessID, careerID int64) error {
	return r.queries.SoftDeleteRecruitCareer(ctx, sqlc.SoftDeleteRecruitCareerParams{
		ID:         careerID,
		BusinessID: businessID,
	})
}

func (r *Repository) GetCustomizationByOwner(ctx context.Context, businessID int64) (domain.Customization, error) {
	row, err := r.queries.GetCustomizationByBusiness(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customization{}, ErrCustomizationNotFound
		}
		return domain.Customization{}, err
	}
	return customizationFromBusinessRow(row), nil
}

func (r *Repository) GetCustomizationBySlug(ctx context.Context, slug string) (domain.Customization, error) {
	row, err := r.queries.GetCustomizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customization{}, ErrCustomizationNotFound
		}
		return domain.Customization{}, err
	}
	return customizationFromSlugRow(row), nil
}

func (r *Repository) UpsertCustomization(ctx context.Context, in domain.CustomizationInput) error {
	return r.queries.UpsertCustomization(ctx, sqlc.UpsertCustomizationParams{
		OwnerID:            in.OwnerID,
		BusinessID:         in.BusinessID,
		Slug:               in.Slug,
		BrandName:          in.BrandName,
		Tagline:            nullStr(in.Tagline),
		AboutBlurb:         nullStr(in.AboutBlurb),
		LogoUrl:            nullStr(in.LogoURL),
		PrimaryColor:       nullStr(in.PrimaryColor),
		AccentColor:        nullStr(in.AccentColor),
		BgColor:            nullStr(in.BgColor),
		TextColor:          nullStr(in.TextColor),
		MutedTextColor:     nullStr(in.MutedTextColor),
		FontPairing:        nullStr(in.FontPairing),
		HeroImageUrl:       nullStr(in.HeroImageURL),
		FaviconUrl:         nullStr(in.FaviconURL),
		Template:           in.Template,
		HeaderHtml:         nullStr(in.HeaderHTML),
		MetaTitle:          nullStr(in.MetaTitle),
		MetaDescription:    nullStr(in.MetaDescription),
		PrivacyUrl:         nullStr(in.PrivacyURL),
		TermsUrl:           nullStr(in.TermsURL),
		GaMeasurementID:    nullStr(in.GAMeasurementID),
		MetaPixelID:        nullStr(in.MetaPixelID),
		RequireConsent:     in.RequireConsent,
		NotificationEmail:  nullStr(in.NotificationEmail),
		EmailNotifications: in.EmailNotifications,
	})
}

func (r *Repository) CreateSubmission(ctx context.Context, in domain.SubmissionInput) (int64, error) {
	result, err := r.queries.CreateCareerSubmission(ctx, sqlc.CreateCareerSubmissionParams{
		CareerID:      in.CareerID,
		OwnerID:       in.OwnerID,
		BusinessID:    in.BusinessID,
		CandidateName: in.CandidateName,
		Email:         in.Email,
		Phone:         nullStr(in.Phone),
		ResumeUrl:     nullStr(in.ResumeURL),
		CoverLetter:   nullStr(in.CoverLetter),
		AnswersJson:   jsonBytesFromString(in.AnswersJSON),
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) ListSubmissions(ctx context.Context, businessID int64, status string) ([]domain.Submission, error) {
	if status != "" {
		rows, err := r.queries.ListCareerSubmissionsByStatus(ctx, sqlc.ListCareerSubmissionsByStatusParams{
			BusinessID: businessID,
			Status:     status,
		})
		if err != nil {
			return nil, err
		}
		return submissionsFromStatusRows(rows), nil
	}

	rows, err := r.queries.ListCareerSubmissions(ctx, businessID)
	if err != nil {
		return nil, err
	}
	return submissionsFromRows(rows), nil
}

func (r *Repository) GetSubmissionResumeURL(ctx context.Context, businessID, submissionID int64) (string, error) {
	return r.queries.GetCareerSubmissionResumeURL(ctx, sqlc.GetCareerSubmissionResumeURLParams{
		ID:         submissionID,
		BusinessID: businessID,
	})
}

func (r *Repository) UpdateSubmissionStatus(ctx context.Context, businessID, submissionID int64, status string) error {
	return r.queries.UpdateCareerSubmissionStatus(ctx, sqlc.UpdateCareerSubmissionStatusParams{
		Status:     status,
		ID:         submissionID,
		BusinessID: businessID,
	})
}

func recruitCareerFromRow(row sqlc.ListRecruitCareersRow) domain.RecruitCareer {
	return domain.RecruitCareer{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		Slug:           row.Slug,
		Title:          row.Title,
		Department:     row.Department,
		Location:       row.Location,
		EmploymentType: row.EmploymentType,
		WorkMode:       row.WorkMode,
		Seniority:      row.Seniority,
		Details:        row.Details,
		Requirements:   row.Requirements,
		SalaryMin:      parseSalary(row.SalaryMin),
		SalaryMax:      parseSalary(row.SalaryMax),
		SalaryCurrency: row.SalaryCurrency,
		IsPublished:    row.IsPublished,
		IsActive:       row.IsActive,
		CreatedAt:      nullTimeValue(row.CreatedAt),
		UpdatedAt:      nullTimeValue(row.UpdatedAt),
	}
}

func recruitCareerFromGetRow(row sqlc.GetRecruitCareerBySlugRow) domain.RecruitCareer {
	return domain.RecruitCareer{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		Slug:           row.Slug,
		Title:          row.Title,
		Department:     row.Department,
		Location:       row.Location,
		EmploymentType: row.EmploymentType,
		WorkMode:       row.WorkMode,
		Seniority:      row.Seniority,
		Details:        row.Details,
		Requirements:   row.Requirements,
		SalaryMin:      parseSalary(row.SalaryMin),
		SalaryMax:      parseSalary(row.SalaryMax),
		SalaryCurrency: row.SalaryCurrency,
		IsPublished:    row.IsPublished,
		IsActive:       row.IsActive,
		CreatedAt:      nullTimeValue(row.CreatedAt),
		UpdatedAt:      nullTimeValue(row.UpdatedAt),
	}
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
