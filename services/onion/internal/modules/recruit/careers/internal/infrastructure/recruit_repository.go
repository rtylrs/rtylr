package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
)

// ErrCareerNotFound / ErrCustomizationNotFound are sentinel errors the
// service layer translates into HTTP 404s.
var (
	ErrCareerNotFound           = domain.ErrCareerNotFound
	ErrSubmissionNotFound       = domain.ErrSubmissionNotFound
	ErrCustomizationNotFound    = domain.ErrCustomizationNotFound
	ErrCustomizationConflict    = errors.New("customization already belongs to another business")
	ErrInvalidSubmissionAnswers = errors.New("invalid application answers")
	ErrRequiredAnswerMissing    = errors.New("required application answer missing")
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

// GetCareerByID reads a single career by primary key inside the tenant. The
// create handler needs the stored row back (slug may be server-generated),
// and id is unambiguous where slug is only unique per owner.
func (r *Repository) GetCareerByID(ctx context.Context, businessID, careerID int64) (domain.RecruitCareer, error) {
	row, err := r.queries.GetRecruitCareerByID(ctx, sqlc.GetRecruitCareerByIDParams{
		ID:         careerID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RecruitCareer{}, ErrCareerNotFound
		}
		return domain.RecruitCareer{}, err
	}
	return recruitCareerFromIDRow(row), nil
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

	result, err := r.queries.UpdateRecruitCareer(ctx, sqlc.UpdateRecruitCareerParams{
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
	if err != nil {
		return err
	}
	return requireRecruitWrite(result, ErrCareerNotFound)
}

func (r *Repository) DeleteCareer(ctx context.Context, businessID, careerID int64) error {
	result, err := r.queries.SoftDeleteRecruitCareer(ctx, sqlc.SoftDeleteRecruitCareerParams{
		ID:         careerID,
		BusinessID: businessID,
	})
	if err != nil {
		return err
	}
	return requireRecruitWrite(result, ErrCareerNotFound)
}

func requireRecruitWrite(result sql.Result, notFound error) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notFound
	}
	return nil
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
	// The table has independent unique keys for owner_id and slug. Lock both
	// scopes before invoking ON DUPLICATE KEY UPDATE so a concurrent first save
	// cannot turn the other tenant's key collision into an update of its row.
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureCustomizationScopeLocked(ctx, tx, in.OwnerID, in.BusinessID, in.Slug); err != nil {
		return err
	}
	if err := sqlc.New(tx).UpsertCustomization(ctx, sqlc.UpsertCustomizationParams{
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
	}); err != nil {
		return err
	}
	return tx.Commit()
}
