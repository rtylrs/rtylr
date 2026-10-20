package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) StartLeadImport(ctx context.Context, req *crmsalesv1.StartLeadImportRequest) (*crmsalesv1.LeadImportResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	batch, err := a.Service.StartImport(ctx, c.businessID, c.userID, req.FileName, int(req.TotalRows))
	if err != nil {
		return nil, fail(ctx, "start lead import", err)
	}
	return &crmsalesv1.LeadImportResponse{Import: importToGen(batch)}, nil
}

func (a *API) ImportLeadRows(ctx context.Context, req *crmsalesv1.ImportLeadRowsRequest) (*crmsalesv1.ImportLeadRowsResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, invalid("rows are required")
	}
	rows := make([]domain.ImportRow, 0, len(req.Body.Rows))
	var rejected []*crmsalesv1.ImportLeadRowResult
	for _, row := range req.Body.Rows {
		if row == nil {
			continue
		}
		lead, err := leadFieldsFromGen(row.Lead)
		if err != nil {
			rejected = append(rejected, &crmsalesv1.ImportLeadRowResult{Row: row.Row, Outcome: string(domain.ImportFailed), Error: "value must be a decimal amount"})
			continue
		}
		rows = append(rows, domain.ImportRow{Row: int(row.Row), Lead: lead})
	}
	results := rejected
	if len(rows) > 0 {
		stored, err := a.Service.ImportLeadRows(ctx, c.businessID, c.userID, req.Id, domain.ImportOptions{
			Duplicates: domain.DuplicatePolicy(req.Body.Duplicates),
			AssignedTo: positivePtr(req.Body.AssignedTo),
			Source:     req.Body.Source,
		}, rows)
		if err != nil {
			return nil, fail(ctx, "import lead rows", err)
		}
		for _, result := range stored {
			out := &crmsalesv1.ImportLeadRowResult{Row: int32(result.Row), Outcome: string(result.Outcome), Error: result.Error} // #nosec G115 -- a file row number.
			if result.LeadID > 0 {
				id := result.LeadID
				out.LeadId = &id
			}
			results = append(results, out)
		}
	}
	return &crmsalesv1.ImportLeadRowsResponse{Results: results}, nil
}

func (a *API) FinishLeadImport(ctx context.Context, req *crmsalesv1.FinishLeadImportRequest) (*crmsalesv1.LeadImportResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	batch, err := a.Service.FinishImport(ctx, c.businessID, req.Id)
	if err != nil {
		return nil, fail(ctx, "finish lead import", err)
	}
	return &crmsalesv1.LeadImportResponse{Import: importToGen(batch)}, nil
}

func (a *API) ListLeadImports(ctx context.Context, _ *crmsalesv1.ListLeadImportsRequest) (*crmsalesv1.ListLeadImportsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	batches, err := a.Service.ListImports(ctx, c.businessID)
	if err != nil {
		return nil, fail(ctx, "list lead imports", err)
	}
	out := make([]*crmsalesv1.LeadImport, 0, len(batches))
	for _, batch := range batches {
		out = append(out, importToGen(batch))
	}
	return &crmsalesv1.ListLeadImportsResponse{Imports: out}, nil
}

func (a *API) UndoLeadImport(ctx context.Context, req *crmsalesv1.UndoLeadImportRequest) (*crmsalesv1.UndoLeadImportResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	result, err := a.Service.UndoImport(ctx, c.businessID, req.Id)
	if err != nil {
		return nil, fail(ctx, "undo lead import", err)
	}
	return &crmsalesv1.UndoLeadImportResponse{Removed: result.Removed, Kept: result.Kept}, nil
}
