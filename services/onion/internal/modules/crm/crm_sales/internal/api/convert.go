package api

import (
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func leadToGen(l domain.Lead) *crmsalesv1.Lead {
	return &crmsalesv1.Lead{
		Id:               l.ID,
		FirstName:        l.FirstName,
		LastName:         l.LastName,
		JobTitle:         l.JobTitle,
		Email:            l.Email,
		Phone:            l.Phone,
		CompanyName:      l.CompanyName,
		Website:          l.Website,
		Linkedin:         l.LinkedIn,
		Industry:         l.Industry,
		Size:             l.Size,
		City:             l.City,
		Country:          l.Country,
		Source:           l.Source,
		Status:           string(l.Status),
		DisqualifyReason: l.DisqualifyReason,
		AssignedTo:       l.AssignedTo,
		Value:            amountPtrString(l.Value),
		Billing:          string(l.Billing),
		Notes:            l.Notes,
		AttemptCount:     int32(l.AttemptCount), // #nosec G115 -- one per logged call.
		LastActivityAt:   unixPtr(l.LastActivityAt),
		NextFollowUp:     stringPtr(l.NextFollowUp),
		CustomerId:       l.CustomerID,
		CompanyId:        l.CompanyID,
		DealId:           l.DealID,
		ConvertedAt:      unixPtr(l.ConvertedAt),
		ImportBatchId:    l.ImportBatchID,
		CreatedAt:        l.CreatedAt.Unix(),
		UpdatedAt:        l.UpdatedAt.Unix(),
	}
}

func leadFieldsFromGen(f *crmsalesv1.LeadFields) (domain.LeadInput, error) {
	if f == nil {
		return domain.LeadInput{}, invalid("lead is required")
	}
	value, err := parseAmountPtr(f.Value, "value")
	if err != nil {
		return domain.LeadInput{}, err
	}
	return domain.LeadInput{
		FirstName:   f.FirstName,
		LastName:    f.LastName,
		JobTitle:    f.JobTitle,
		Email:       f.Email,
		Phone:       f.Phone,
		CompanyName: f.CompanyName,
		Website:     f.Website,
		LinkedIn:    f.Linkedin,
		Industry:    f.Industry,
		Size:        f.Size,
		City:        f.City,
		Country:     f.Country,
		Source:      f.Source,
		AssignedTo:  positivePtr(f.AssignedTo),
		Value:       value,
		Billing:     domain.Billing(f.Billing),
		Notes:       f.Notes,
	}, nil
}

func dealToGen(d domain.Deal) *crmsalesv1.Deal {
	return &crmsalesv1.Deal{
		Id:                d.ID,
		PipelineId:        d.PipelineID,
		StageId:           d.StageID,
		Title:             d.Title,
		LeadId:            d.LeadID,
		CustomerId:        d.CustomerID,
		CompanyId:         d.CompanyID,
		ContactName:       d.ContactName,
		CompanyName:       d.CompanyName,
		Value:             d.Value.String(),
		Billing:           string(d.Billing),
		Amount:            d.Amount().String(),
		ExpectedCloseDate: stringPtr(d.ExpectedCloseDate),
		AssignedTo:        d.AssignedTo,
		Source:            d.Source,
		Status:            string(d.Status),
		LostReason:        d.LostReason,
		ClosedAt:          unixPtr(d.ClosedAt),
		Notes:             d.Notes,
		LastActivityAt:    unixPtr(d.LastActivityAt),
		NextFollowUp:      stringPtr(d.NextFollowUp),
		StageChangedAt:    d.StageChangedAt.Unix(),
		CreatedAt:         d.CreatedAt.Unix(),
	}
}

func dealsToGen(deals []domain.Deal) []*crmsalesv1.Deal {
	out := make([]*crmsalesv1.Deal, 0, len(deals))
	for _, deal := range deals {
		out = append(out, dealToGen(deal))
	}
	return out
}

func dealFieldsFromGen(f *crmsalesv1.DealFields) (domain.DealInput, error) {
	if f == nil {
		return domain.DealInput{}, nil
	}
	value, err := parseAmount(f.Value, "value")
	if err != nil {
		return domain.DealInput{}, err
	}
	return domain.DealInput{
		Title:             f.Title,
		PipelineID:        int64Value(f.PipelineId),
		StageID:           int64Value(f.StageId),
		ContactName:       f.ContactName,
		CompanyName:       f.CompanyName,
		Value:             value,
		Billing:           domain.Billing(f.Billing),
		ExpectedCloseDate: f.ExpectedCloseDate,
		AssignedTo:        positivePtr(f.AssignedTo),
		Source:            f.Source,
		Notes:             f.Notes,
	}, nil
}

func activityToGen(a domain.Activity) *crmsalesv1.Activity {
	activity := &crmsalesv1.Activity{
		Id:         a.ID,
		LeadId:     a.LeadID,
		DealId:     a.DealID,
		Type:       string(a.Type),
		Outcome:    string(a.Outcome),
		Notes:      a.Notes,
		CreatedBy:  a.CreatedBy,
		OccurredAt: a.OccurredAt.Unix(),
	}
	if a.DurationSeconds != nil {
		seconds := int32(*a.DurationSeconds) // #nosec G115 -- capped at a day.
		activity.DurationSeconds = &seconds
	}
	return activity
}

func activityPageToGen(page domain.ActivityPage) *crmsalesv1.ListActivitiesResponse {
	out := make([]*crmsalesv1.Activity, 0, len(page.Activities))
	for _, activity := range page.Activities {
		out = append(out, activityToGen(activity))
	}
	return &crmsalesv1.ListActivitiesResponse{Activities: out, NextCursor: page.NextCursor}
}

func taskToGen(t domain.Task) *crmsalesv1.SalesTask {
	return &crmsalesv1.SalesTask{
		Id:          t.ID,
		LeadId:      t.LeadID,
		DealId:      t.DealID,
		Title:       t.Title,
		Type:        string(t.Type),
		DueDate:     t.DueDate,
		AssignedTo:  t.AssignedTo,
		CompletedAt: unixPtr(t.CompletedAt),
		LeadName:    t.LeadName,
		CompanyName: t.CompanyName,
		Phone:       t.Phone,
		DealTitle:   t.DealTitle,
		CreatedAt:   t.CreatedAt.Unix(),
	}
}

func taskFieldsFromGen(f *crmsalesv1.SalesTaskFields) (domain.TaskInput, error) {
	if f == nil {
		return domain.TaskInput{}, invalid("task is required")
	}
	return domain.TaskInput{
		LeadID:     positivePtr(f.LeadId),
		DealID:     positivePtr(f.DealId),
		Title:      f.Title,
		Type:       domain.TaskType(f.Type),
		DueDate:    f.DueDate,
		AssignedTo: positivePtr(f.AssignedTo),
	}, nil
}

func stageToGen(s domain.Stage) *crmsalesv1.Stage {
	return &crmsalesv1.Stage{
		Id:          s.ID,
		PipelineId:  s.PipelineID,
		Name:        s.Name,
		Position:    int32(s.Position),    // #nosec G115 -- at most 20 stages.
		Probability: int32(s.Probability), // #nosec G115 -- 0-100.
		Kind:        string(s.Kind),
	}
}

func pipelineToGen(p domain.Pipeline) *crmsalesv1.Pipeline {
	stages := make([]*crmsalesv1.Stage, 0, len(p.Stages))
	for _, stage := range p.Stages {
		stages = append(stages, stageToGen(stage))
	}
	return &crmsalesv1.Pipeline{Id: p.ID, Name: p.Name, IsDefault: p.IsDefault, Stages: stages}
}

func pipelineFieldsFromGen(f *crmsalesv1.PipelineFields) (domain.PipelineInput, error) {
	if f == nil {
		return domain.PipelineInput{}, invalid("pipeline is required")
	}
	in := domain.PipelineInput{Name: f.Name}
	for _, stage := range f.Stages {
		if stage == nil {
			continue
		}
		in.Stages = append(in.Stages, domain.StageInput{
			ID:          positivePtr(stage.Id),
			Name:        stage.Name,
			Probability: int(stage.Probability),
			Kind:        domain.StageKind(stage.Kind),
		})
	}
	return in, nil
}

func contractToGen(c domain.Contract) *crmsalesv1.Contract {
	return &crmsalesv1.Contract{
		Id:           c.ID,
		LeadId:       c.LeadID,
		DealId:       c.DealID,
		Plan:         c.Plan,
		Mrr:          c.MRR.String(),
		Seats:        int32(c.Seats), // #nosec G115 -- seat counts are small.
		Value:        c.Value.String(),
		Billing:      string(c.Billing),
		StartDate:    c.StartDate,
		EndDate:      c.EndDate,
		CustomerId:   c.CustomerID,
		SignedAt:     c.SignedAt,
		SignedByName: c.SignedByName,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
	}
}

func contractPtrToGen(c *domain.Contract) *crmsalesv1.Contract {
	if c == nil {
		return nil
	}
	return contractToGen(*c)
}

func contractFieldsFromGen(f *crmsalesv1.ContractFields) (*domain.ContractInput, error) {
	if f == nil {
		return nil, nil
	}
	value, err := parseAmount(f.Value, "value")
	if err != nil {
		return nil, err
	}
	return &domain.ContractInput{
		Plan:         f.Plan,
		Seats:        int(f.Seats),
		Value:        value,
		Billing:      domain.Billing(f.Billing),
		StartDate:    f.StartDate,
		EndDate:      f.EndDate,
		SignedAt:     f.SignedAt,
		SignedByName: f.SignedByName,
		Notes:        f.Notes,
	}, nil
}

func importToGen(b domain.ImportBatch) *crmsalesv1.LeadImport {
	return &crmsalesv1.LeadImport{
		Id:          b.ID,
		FileName:    b.FileName,
		TotalRows:   int32(b.TotalRows), // #nosec G115 -- capped at 200,000.
		Created:     int32(b.Created),   // #nosec G115 -- bounded by total rows.
		Updated:     int32(b.Updated),   // #nosec G115 -- bounded by total rows.
		Skipped:     int32(b.Skipped),   // #nosec G115 -- bounded by total rows.
		Failed:      int32(b.Failed),    // #nosec G115 -- bounded by total rows.
		Status:      b.Status,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt.Unix(),
		CompletedAt: unixPtr(b.CompletedAt),
	}
}
