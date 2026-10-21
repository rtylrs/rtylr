package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// MaxImportFileRows bounds one import, so a mistaken upload cannot run away.
const MaxImportFileRows = 200000

func (s *Service) StartImport(ctx context.Context, businessID, actorID int64, fileName string, totalRows int) (domain.ImportBatch, error) {
	if totalRows <= 0 || totalRows > MaxImportFileRows {
		return domain.ImportBatch{}, &domain.ValidationError{Message: "an import needs between 1 and 200,000 rows"}
	}
	if len([]rune(fileName)) > 255 {
		fileName = string([]rune(fileName)[:255])
	}
	return s.repo.CreateImportBatch(ctx, businessID, actorID, fileName, totalRows)
}

// ImportLeadRows stores one chunk of a running import. Rows that fail
// validation are reported back with the reason instead of failing the chunk.
func (s *Service) ImportLeadRows(ctx context.Context, businessID, actorID, batchID int64, opts domain.ImportOptions, rows []domain.ImportRow) ([]domain.ImportRowResult, error) {
	if len(rows) == 0 || len(rows) > domain.MaxImportRows {
		return nil, &domain.ValidationError{Message: "send between 1 and 500 rows at a time"}
	}
	if opts.Duplicates == "" {
		opts.Duplicates = domain.DuplicatesSkip
	}
	if opts.Duplicates != domain.DuplicatesSkip && opts.Duplicates != domain.DuplicatesFill {
		return nil, &domain.ValidationError{Message: "duplicates must be skip or fill"}
	}
	if err := s.validateAssignee(ctx, businessID, opts.AssignedTo); err != nil {
		return nil, err
	}
	valid := make([]domain.ImportRow, 0, len(rows))
	var failed []domain.ImportRowResult
	for _, row := range rows {
		if row.Lead.Source == "" {
			row.Lead.Source = opts.Source
		}
		if row.Lead.AssignedTo == nil {
			row.Lead.AssignedTo = opts.AssignedTo
		} else if err := s.validateAssignee(ctx, businessID, row.Lead.AssignedTo); err != nil {
			row.Lead.AssignedTo = opts.AssignedTo
		}
		if err := row.Lead.Normalize(); err != nil {
			failed = append(failed, domain.ImportRowResult{Row: row.Row, Outcome: domain.ImportFailed, Error: err.Error()})
			continue
		}
		valid = append(valid, row)
	}
	return s.repo.ImportLeadRows(ctx, businessID, actorID, batchID, opts, valid, failed)
}

func (s *Service) FinishImport(ctx context.Context, businessID, id int64) (domain.ImportBatch, error) {
	return s.repo.FinishImport(ctx, businessID, id, s.now())
}

func (s *Service) ListImports(ctx context.Context, businessID int64) ([]domain.ImportBatch, error) {
	return s.repo.ListImportBatches(ctx, businessID)
}

func (s *Service) UndoImport(ctx context.Context, businessID, id int64) (domain.UndoImportResult, error) {
	return s.repo.UndoImport(ctx, businessID, id, s.now())
}
