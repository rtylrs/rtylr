package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
)

func ensureCustomizationScopeLocked(ctx context.Context, tx *sql.Tx, ownerID, businessID int64, slug string) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT business_id
		FROM careers_customizations
		WHERE slug = ? OR owner_id = ?
		FOR UPDATE`, slug, ownerID)
	if err != nil {
		return err
	}
	defer rows.Close()

	conflict := false
	for rows.Next() {
		var existingBusinessID int64
		if err := rows.Scan(&existingBusinessID); err != nil {
			return err
		}
		if existingBusinessID != businessID {
			conflict = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if conflict {
		return ErrCustomizationConflict
	}
	return nil
}

func (r *Repository) CreateSubmission(ctx context.Context, in domain.SubmissionInput) (int64, error) {
	if err := r.validateRequiredSubmissionAnswers(ctx, in.CareerID, in.AnswersJSON); err != nil {
		return 0, err
	}

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

func (r *Repository) validateRequiredSubmissionAnswers(ctx context.Context, careerID int64, answersJSON string) error {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id
		FROM careers_form_fields
		WHERE career_id = ? AND required = true`, careerID)
	if err != nil {
		return err
	}
	defer rows.Close()

	answers := make(map[string]json.RawMessage)
	if strings.TrimSpace(answersJSON) != "" {
		if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
			return ErrInvalidSubmissionAnswers
		}
	}
	for rows.Next() {
		var fieldID int64
		if err := rows.Scan(&fieldID); err != nil {
			return err
		}
		answer, ok := answers[strconv.FormatInt(fieldID, 10)]
		if !ok || emptySubmissionAnswer(answer) {
			return ErrRequiredAnswerMissing
		}
	}
	return rows.Err()
}

func emptySubmissionAnswer(answer json.RawMessage) bool {
	var value any
	if err := json.Unmarshal(answer, &value); err != nil {
		return true
	}
	switch value := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(value) == ""
	case []any:
		return len(value) == 0
	case map[string]any:
		return len(value) == 0
	default:
		return false
	}
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

// ListSubmissionsPage is the paged form of ListSubmissions for callers that
// must see every row rather than the first UI page (e.g. the CSV export). Rows
// are ordered newest-first with id as the tie-breaker so OFFSET paging is
// stable when rows share a created_at.
func (r *Repository) ListSubmissionsPage(ctx context.Context, businessID int64, status string, limit, offset int32) ([]domain.Submission, error) {
	if status != "" {
		rows, err := r.queries.ListCareerSubmissionsByStatusPage(ctx, sqlc.ListCareerSubmissionsByStatusPageParams{
			BusinessID: businessID,
			Status:     status,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		return submissionsFromStatusPageRows(rows), nil
	}

	rows, err := r.queries.ListCareerSubmissionsPage(ctx, sqlc.ListCareerSubmissionsPageParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	return submissionsFromPageRows(rows), nil
}

func (r *Repository) GetSubmissionResumeURL(ctx context.Context, businessID, submissionID int64) (string, error) {
	resumeURL, err := r.queries.GetCareerSubmissionResumeURL(ctx, sqlc.GetCareerSubmissionResumeURLParams{
		ID:         submissionID,
		BusinessID: businessID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrSubmissionNotFound
	}
	return resumeURL, err
}

func (r *Repository) UpdateSubmissionStatus(ctx context.Context, businessID, submissionID int64, status string) error {
	result, err := r.queries.UpdateCareerSubmissionStatus(ctx, sqlc.UpdateCareerSubmissionStatusParams{
		Status:     status,
		ID:         submissionID,
		BusinessID: businessID,
	})
	if err != nil {
		return err
	}
	return requireRecruitWrite(result, ErrSubmissionNotFound)
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
