package infrastructure

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) CreateImportBatch(ctx context.Context, businessID, actorID int64, fileName string, totalRows int) (domain.ImportBatch, error) {
	result, err := r.queries.CreateImportBatch(ctx, sqlc.CreateImportBatchParams{
		BusinessID: businessID, CreatedBy: actorID, FileName: fileName, TotalRows: int32(totalRows), // #nosec G115 -- validated by the service.
	})
	if err != nil {
		return domain.ImportBatch{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.ImportBatch{}, err
	}
	return r.GetImportBatch(ctx, businessID, id)
}

func (r *Repository) GetImportBatch(ctx context.Context, businessID, id int64) (domain.ImportBatch, error) {
	row, err := r.queries.GetImportBatch(ctx, sqlc.GetImportBatchParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.ImportBatch{}, notFound(err, domain.ErrImportNotFound)
	}
	return importBatchFromRow(row), nil
}

func (r *Repository) ListImportBatches(ctx context.Context, businessID int64) ([]domain.ImportBatch, error) {
	rows, err := r.queries.ListImportBatches(ctx, businessID)
	if err != nil {
		return nil, err
	}
	batches := make([]domain.ImportBatch, 0, len(rows))
	for _, row := range rows {
		batches = append(batches, importBatchFromRow(row))
	}
	return batches, nil
}

// importMatcher finds the existing lead a row duplicates: by email, then
// phone, then company plus person name. Rows created earlier in the same
// chunk are matched too.
type importMatcher struct {
	byEmail   map[string]int64
	byPhone   map[string]int64
	byCompany map[string]int64
}

func companyKey(company, first, last string) string {
	return strings.ToLower(company) + "\x00" + strings.ToLower(strings.TrimSpace(first+" "+last))
}

func (m importMatcher) match(in domain.LeadInput) int64 {
	if id, ok := m.byEmail[in.Email]; in.Email != "" && ok {
		return id
	}
	if digits := domain.PhoneDigits(in.Phone); digits != "" {
		if id, ok := m.byPhone[digits]; ok {
			return id
		}
	}
	if in.CompanyName != "" {
		if id, ok := m.byCompany[companyKey(in.CompanyName, in.FirstName, in.LastName)]; ok {
			return id
		}
	}
	return 0
}

func (m importMatcher) remember(id int64, in domain.LeadInput) {
	if in.Email != "" {
		m.byEmail[in.Email] = id
	}
	if digits := domain.PhoneDigits(in.Phone); digits != "" {
		m.byPhone[digits] = id
	}
	if in.CompanyName != "" {
		m.byCompany[companyKey(in.CompanyName, in.FirstName, in.LastName)] = id
	}
}

func loadImportMatcher(ctx context.Context, q *sqlc.Queries, businessID int64, rows []domain.ImportRow) (importMatcher, error) {
	m := importMatcher{byEmail: map[string]int64{}, byPhone: map[string]int64{}, byCompany: map[string]int64{}}
	var emails, phones, companies []string
	for _, row := range rows {
		if row.Lead.Email != "" {
			emails = append(emails, row.Lead.Email)
		}
		if digits := domain.PhoneDigits(row.Lead.Phone); digits != "" {
			phones = append(phones, digits)
		}
		if row.Lead.CompanyName != "" {
			companies = append(companies, row.Lead.CompanyName)
		}
	}
	if len(emails) > 0 {
		found, err := q.FindLeadsByEmail(ctx, sqlc.FindLeadsByEmailParams{BusinessID: businessID, Emails: nullStrings(emails)})
		if err != nil {
			return m, err
		}
		for _, lead := range found {
			m.byEmail[strings.ToLower(lead.Email.String)] = lead.ID
		}
	}
	if len(phones) > 0 {
		found, err := q.FindLeadsByPhone(ctx, sqlc.FindLeadsByPhoneParams{BusinessID: businessID, Phones: nullStrings(phones)})
		if err != nil {
			return m, err
		}
		for _, lead := range found {
			m.byPhone[lead.PhoneDigits.String] = lead.ID
		}
	}
	if len(companies) > 0 {
		found, err := q.FindLeadsByCompany(ctx, sqlc.FindLeadsByCompanyParams{BusinessID: businessID, Companies: companies})
		if err != nil {
			return m, err
		}
		for _, lead := range found {
			m.byCompany[companyKey(lead.CompanyName, lead.FirstName.String, lead.LastName.String)] = lead.ID
		}
	}
	return m, nil
}

func nullStrings(values []string) []sql.NullString {
	out := make([]sql.NullString, 0, len(values))
	for _, value := range values {
		out = append(out, sql.NullString{String: value, Valid: true})
	}
	return out
}

// ImportLeadRows stores one chunk of normalized rows in a running import.
// failed holds rows the service already rejected, counted with the rest.
func (r *Repository) ImportLeadRows(ctx context.Context, businessID, actorID, batchID int64, opts domain.ImportOptions, rows []domain.ImportRow, failed []domain.ImportRowResult) ([]domain.ImportRowResult, error) {
	results := append([]domain.ImportRowResult{}, failed...)
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		batch, err := q.GetImportBatchForUpdate(ctx, sqlc.GetImportBatchForUpdateParams{ID: batchID, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrImportNotFound)
		}
		if batch.Status != sqlc.CrmImportBatchesStatusRunning {
			return domain.ErrImportClosed
		}
		matcher, err := loadImportMatcher(ctx, q, businessID, rows)
		if err != nil {
			return err
		}
		var created, updated, skipped int32
		for _, row := range rows {
			in := row.Lead
			if existing := matcher.match(in); existing > 0 {
				if opts.Duplicates != domain.DuplicatesFill {
					skipped++
					results = append(results, domain.ImportRowResult{Row: row.Row, Outcome: domain.ImportSkipped, LeadID: existing})
					continue
				}
				if err := q.FillLeadBlanks(ctx, sqlc.FillLeadBlanksParams{
					FirstName: nullString(in.FirstName), LastName: nullString(in.LastName), JobTitle: nullString(in.JobTitle),
					Email: nullString(in.Email), Phone: nullString(in.Phone), PhoneDigits: nullString(domain.PhoneDigits(in.Phone)),
					CompanyName: in.CompanyName, Website: nullString(in.Website), Linkedin: nullString(in.LinkedIn),
					Industry: nullString(in.Industry), Size: nullString(in.Size), City: nullString(in.City),
					Country: nullString(in.Country), Value: nullAmount(in.Value), Notes: nullString(in.Notes),
					ID: existing, BusinessID: businessID,
				}); err != nil {
					return err
				}
				updated++
				results = append(results, domain.ImportRowResult{Row: row.Row, Outcome: domain.ImportUpdated, LeadID: existing})
				continue
			}
			result, err := q.CreateLead(ctx, createLeadParams(businessID, actorID, in, &batchID))
			if err != nil {
				return err
			}
			id, err := result.LastInsertId()
			if err != nil {
				return err
			}
			matcher.remember(id, in)
			created++
			results = append(results, domain.ImportRowResult{Row: row.Row, Outcome: domain.ImportCreated, LeadID: id})
		}
		return q.AddImportBatchCounts(ctx, sqlc.AddImportBatchCountsParams{
			CreatedCount: created, UpdatedCount: updated, SkippedCount: skipped,
			FailedCount: int32(len(failed)), // #nosec G115 -- at most one chunk.
			ID:          batchID, BusinessID: businessID,
		})
	})
	return results, err
}

// FinishImport marks a running import completed.
func (r *Repository) FinishImport(ctx context.Context, businessID, id int64, now time.Time) (domain.ImportBatch, error) {
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		batch, err := q.GetImportBatchForUpdate(ctx, sqlc.GetImportBatchForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrImportNotFound)
		}
		if batch.Status != sqlc.CrmImportBatchesStatusRunning {
			return nil
		}
		return q.SetImportBatchStatus(ctx, sqlc.SetImportBatchStatusParams{
			Status: sqlc.CrmImportBatchesStatusCompleted, CompletedAt: someTime(now), ID: id, BusinessID: businessID,
		})
	})
	if err != nil {
		return domain.ImportBatch{}, err
	}
	return r.GetImportBatch(ctx, businessID, id)
}

// UndoImport removes the import's leads nobody has worked since and keeps the
// rest.
func (r *Repository) UndoImport(ctx context.Context, businessID, id int64, now time.Time) (domain.UndoImportResult, error) {
	var result domain.UndoImportResult
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		batch, err := q.GetImportBatchForUpdate(ctx, sqlc.GetImportBatchForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrImportNotFound)
		}
		if batch.Status == sqlc.CrmImportBatchesStatusUndone {
			return domain.ErrImportClosed
		}
		deleted, err := q.DeleteUntouchedBatchLeads(ctx, sqlc.DeleteUntouchedBatchLeadsParams{BusinessID: businessID, ImportBatchID: someInt64(id)})
		if err != nil {
			return err
		}
		if result.Removed, err = deleted.RowsAffected(); err != nil {
			return err
		}
		if result.Kept, err = q.CountBatchLeads(ctx, sqlc.CountBatchLeadsParams{BusinessID: businessID, ImportBatchID: someInt64(id)}); err != nil {
			return err
		}
		return q.SetImportBatchStatus(ctx, sqlc.SetImportBatchStatusParams{
			Status: sqlc.CrmImportBatchesStatusUndone, CompletedAt: someTime(now), ID: id, BusinessID: businessID,
		})
	})
	return result, err
}
