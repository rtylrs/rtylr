package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/internal/outbox"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

const (
	business = testutil.OwnerID
	actor    = testutil.OwnerID
)

var clock = time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)

func newService(t *testing.T) (*application.Service, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)
	return application.NewService(context.Background(), configuredRepository(db)).WithClock(func() time.Time { return clock }), db
}

// unique keeps rows from different tests and runs apart in the shared fixture
// business.
func unique(t *testing.T, label string) string {
	t.Helper()
	return fmt.Sprintf("%s %s %d", label, t.Name(), time.Now().UnixNano())
}

func today() domain.CallerDay {
	return domain.NewCallerDay(clock.Format(time.DateOnly), 0, clock)
}

func TestLeadIsWorkedConvertedAndWon(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()
	company := unique(t, "Acme")

	lead, err := svc.CreateLead(ctx, business, actor, domain.LeadInput{
		FirstName: "Jane", LastName: "Doe", JobTitle: "CEO", Email: "Jane.Doe@Example.com",
		Phone: "+961 3 123 456", CompanyName: company, Source: "Trade show",
		Value: ptr(money.Amount(50000)), Billing: domain.BillingMonthly,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.Status != domain.LeadNew || lead.Email != "jane.doe@example.com" {
		t.Fatalf("created lead = %+v", lead)
	}

	call := func(outcome domain.CallOutcome, followUp *domain.TaskInput) {
		t.Helper()
		if _, _, err := svc.LogActivity(ctx, business, actor, domain.ActivityInput{
			LeadID: &lead.ID, Type: domain.ActivityCall, Outcome: outcome, FollowUp: followUp,
		}); err != nil {
			t.Fatalf("LogActivity(%s): %v", outcome, err)
		}
	}
	call(domain.CallNoAnswer, nil)
	if lead, _ = svc.GetLead(ctx, business, lead.ID); lead.Status != domain.LeadAttempting || lead.AttemptCount != 1 {
		t.Fatalf("after no answer: status %s, attempts %d", lead.Status, lead.AttemptCount)
	}
	call(domain.CallCallback, &domain.TaskInput{Type: domain.TaskCall, Title: "Call back", DueDate: "2026-10-22"})
	lead, _ = svc.GetLead(ctx, business, lead.ID)
	if lead.Status != domain.LeadConnected || lead.AttemptCount != 2 || lead.NextFollowUp != "2026-10-22" {
		t.Fatalf("after callback: status %s, attempts %d, follow-up %q", lead.Status, lead.AttemptCount, lead.NextFollowUp)
	}

	timeline, err := svc.ListLeadActivities(ctx, business, lead.ID, "", 10)
	if err != nil || len(timeline.Activities) != 2 || timeline.Activities[0].Outcome != domain.CallCallback {
		t.Fatalf("timeline = %+v, %v", timeline, err)
	}

	converted, deal, err := svc.ConvertLead(ctx, business, actor, lead.ID, domain.ConvertLeadInput{CreateDeal: true})
	if err != nil {
		t.Fatalf("ConvertLead: %v", err)
	}
	if converted.Status != domain.LeadQualified || converted.DealID == nil || converted.CompanyID == nil || deal == nil {
		t.Fatalf("converted = %+v, deal = %+v", converted, deal)
	}
	if deal.Title != company || deal.ContactName != "Jane Doe" || deal.Value != 50000 || deal.Billing != domain.BillingMonthly {
		t.Fatalf("deal did not inherit the lead: %+v", deal)
	}
	if deal.NextFollowUp != "2026-10-22" {
		t.Fatalf("the lead's open task should carry to the deal, follow-up = %q", deal.NextFollowUp)
	}
	pipeline, columns, err := svc.DealBoard(ctx, business, 0, nil, company)
	if err != nil || !pipeline.IsDefault || len(columns) != 6 || columns[0].Stage.ID != deal.StageID || columns[0].Deals != 1 {
		t.Fatalf("board = %+v %+v, %v", pipeline, columns, err)
	}
	if columns[0].Amount != 600000 || columns[0].WeightedAmount != 60000 {
		t.Fatalf("a monthly 500.00 deal at 10%% is 6000.00 a year, 600.00 weighted: %+v", columns[0])
	}
	if !publishedContact(t, db, converted.ID) {
		t.Fatal("conversion did not publish the lead's contact for the customers module")
	}
	if _, _, err := svc.ConvertLead(ctx, business, actor, lead.ID, domain.ConvertLeadInput{}); !errors.Is(err, domain.ErrLeadConverted) {
		t.Fatalf("second conversion: err = %v, want ErrLeadConverted", err)
	}
	if err := svc.DeleteLead(ctx, business, lead.ID); !errors.Is(err, domain.ErrLeadConverted) {
		t.Fatalf("deleting a converted lead: err = %v, want ErrLeadConverted", err)
	}

	dealTimeline, err := svc.ListDealActivities(ctx, business, deal.ID, "", 10)
	if err != nil || len(dealTimeline.Activities) != 2 {
		t.Fatalf("the deal's timeline should include the lead's calls: %+v, %v", dealTimeline, err)
	}

	won, contract, err := svc.WinDeal(ctx, business, actor, deal.ID, domain.WinDealInput{Contract: &domain.ContractInput{}})
	if err != nil {
		t.Fatalf("WinDeal: %v", err)
	}
	if won.Status != domain.DealWon || won.ClosedAt == nil || contract == nil {
		t.Fatalf("won = %+v, contract = %+v", won, contract)
	}
	if contract.DealID == nil || *contract.DealID != deal.ID || contract.LeadID == nil || *contract.LeadID != lead.ID ||
		contract.Plan != company || contract.Value != 50000 || contract.Billing != domain.BillingMonthly || contract.MRR != 50000 {
		t.Fatalf("contract did not come from the deal: %+v", contract)
	}
	if contract.StartDate == nil || *contract.StartDate != "2026-10-20" || contract.SignedAt != nil {
		t.Fatalf("a drafted contract starts today and is unsigned: %+v", contract)
	}
	if err := svc.DeleteDeal(ctx, business, deal.ID); !errors.Is(err, domain.ErrDealHasContract) {
		t.Fatalf("deleting a deal with a contract: err = %v, want ErrDealHasContract", err)
	}
	if _, err := svc.LoseDeal(ctx, business, actor, deal.ID, ""); !domain.IsValidation(err) {
		t.Fatalf("losing without a reason: err = %v, want a validation error", err)
	}
	reopened, err := svc.ReopenDeal(ctx, business, actor, deal.ID, 0)
	if err != nil || reopened.Status != domain.DealOpen || reopened.ClosedAt != nil {
		t.Fatalf("reopened = %+v, %v", reopened, err)
	}
}

func publishedContact(t *testing.T, db *sql.DB, leadID int64) bool {
	t.Helper()
	pending, err := outbox.NewSource(db).Pending(context.Background(), 1000)
	if err != nil {
		t.Fatalf("pending facts: %v", err)
	}
	for _, fact := range pending {
		if fact.SubjectType == "crm_lead" && fact.SubjectID == leadID {
			return true
		}
	}
	return false
}

func TestLeadListPagesFiltersAndSorts(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	batch := unique(t, "Paging")
	var ids []int64
	for i := 0; i < 7; i++ {
		lead, err := svc.CreateLead(ctx, business, actor, domain.LeadInput{
			CompanyName: fmt.Sprintf("%s %c", batch, 'G'-rune(i)),
			Phone:       fmt.Sprintf("+1 555 0100 %03d", i),
			Source:      batch,
		})
		if err != nil {
			t.Fatalf("CreateLead %d: %v", i, err)
		}
		ids = append(ids, lead.ID)
	}
	if _, err := svc.CreateTask(ctx, business, actor, domain.TaskInput{LeadID: &ids[2], Type: domain.TaskCall, DueDate: "2026-10-19"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := svc.CreateTask(ctx, business, actor, domain.TaskInput{LeadID: &ids[4], Type: domain.TaskCall, DueDate: "2026-10-25"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	var seen []int64
	cursor := ""
	for page := 0; page < 5; page++ {
		result, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Limit: 3, Cursor: cursor}, today())
		if err != nil {
			t.Fatalf("ListLeads page %d: %v", page, err)
		}
		if result.Total != 7 {
			t.Fatalf("total = %d, want 7", result.Total)
		}
		for _, lead := range result.Leads {
			seen = append(seen, lead.ID)
		}
		if cursor = result.NextCursor; cursor == "" {
			break
		}
	}
	if len(seen) != 7 {
		t.Fatalf("paged through %d leads, want 7: %v", len(seen), seen)
	}
	for i := range seen {
		if seen[i] != ids[6-i] {
			t.Fatalf("newest first: got %v, want the reverse of %v", seen, ids)
		}
	}

	byCompany, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Sort: domain.LeadSortCompany, Limit: 2}, today())
	if err != nil || byCompany.Leads[0].ID != ids[6] || byCompany.Leads[1].ID != ids[5] {
		t.Fatalf("company A-Z: %+v, %v", byCompany.Leads, err)
	}
	byFollowUp, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Sort: domain.LeadSortFollowUp, Limit: 3}, today())
	if err != nil || byFollowUp.Leads[0].ID != ids[2] || byFollowUp.Leads[1].ID != ids[4] {
		t.Fatalf("soonest follow-up first: %+v, %v", byFollowUp.Leads, err)
	}
	if _, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Sort: domain.LeadSortCompany, Cursor: byFollowUp.NextCursor}, today()); !errors.Is(err, domain.ErrInvalidCursor) {
		t.Fatalf("a cursor from another sort: err = %v, want ErrInvalidCursor", err)
	}

	overdue, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, FollowUp: domain.FollowUpOverdue}, today())
	if err != nil || overdue.Total != 1 || overdue.Leads[0].ID != ids[2] {
		t.Fatalf("overdue: %+v, %v", overdue, err)
	}
	byPhone, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Search: "0100-003"}, today())
	if err != nil || byPhone.Total != 1 || byPhone.Leads[0].ID != ids[3] {
		t.Fatalf("search by phone digits: %+v, %v", byPhone, err)
	}

	updated, err := svc.UpdateLeads(ctx, business, domain.BulkLeadUpdate{IDs: ids[:3], Status: domain.LeadDisqualified, DisqualifyReason: "Out of area"})
	if err != nil || updated != 3 {
		t.Fatalf("UpdateLeads = %d, %v", updated, err)
	}
	disqualified, err := svc.ListLeads(ctx, business, domain.LeadFilter{Source: batch, Statuses: []domain.LeadStatus{domain.LeadDisqualified}}, today())
	if err != nil || disqualified.Total != 3 || disqualified.Leads[0].DisqualifyReason != "Out of area" {
		t.Fatalf("disqualified: %+v, %v", disqualified, err)
	}
	deleted, err := svc.DeleteLeads(ctx, business, ids)
	if err != nil || deleted != 7 {
		t.Fatalf("DeleteLeads = %d, %v", deleted, err)
	}
}

func TestImportDedupesFillsAndUndoes(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	company := unique(t, "Import Co")
	email := fmt.Sprintf("buyer+%d@example.com", time.Now().UnixNano())
	phone := fmt.Sprintf("+1 (555) %07d", time.Now().UnixNano()%10000000)

	batch, err := svc.StartImport(ctx, business, actor, "list.csv", 7)
	if err != nil {
		t.Fatalf("StartImport: %v", err)
	}
	rows := []domain.ImportRow{
		{Row: 2, Lead: domain.LeadInput{FirstName: "Ana", Email: email}},
		{Row: 3, Lead: domain.LeadInput{FirstName: "Ana again", Email: "  " + email + " "}},
		{Row: 4, Lead: domain.LeadInput{FirstName: "Bo", Phone: phone}},
		{Row: 5, Lead: domain.LeadInput{FirstName: "Bo again", Phone: domain.PhoneDigits(phone)}},
		{Row: 6, Lead: domain.LeadInput{FirstName: "Cy", LastName: "Lee", CompanyName: company}},
		{Row: 7, Lead: domain.LeadInput{Email: "nobody@example.com"}},
		{Row: 8, Lead: domain.LeadInput{FirstName: "Di", Email: "not an email"}},
	}
	results, err := svc.ImportLeadRows(ctx, business, actor, batch.ID, domain.ImportOptions{Source: "Spring list"}, rows)
	if err != nil {
		t.Fatalf("ImportLeadRows: %v", err)
	}
	outcomes := map[int]domain.ImportOutcome{}
	for _, result := range results {
		outcomes[result.Row] = result.Outcome
	}
	want := map[int]domain.ImportOutcome{2: "created", 3: "skipped", 4: "created", 5: "skipped", 6: "created", 7: "failed", 8: "failed"}
	for row, outcome := range want {
		if outcomes[row] != outcome {
			t.Fatalf("row %d: %s, want %s (all: %v)", row, outcomes[row], outcome, outcomes)
		}
	}

	fill, err := svc.ImportLeadRows(ctx, business, actor, batch.ID, domain.ImportOptions{Duplicates: domain.DuplicatesFill}, []domain.ImportRow{
		{Row: 9, Lead: domain.LeadInput{FirstName: "Cy", LastName: "Lee", CompanyName: company, JobTitle: "Buyer"}},
	})
	if err != nil || fill[0].Outcome != domain.ImportUpdated {
		t.Fatalf("fill: %+v, %v", fill, err)
	}
	cy, err := svc.GetLead(ctx, business, fill[0].LeadID)
	if err != nil || cy.JobTitle != "Buyer" || cy.Source != "Spring list" {
		t.Fatalf("filled lead = %+v, %v", cy, err)
	}

	finished, err := svc.FinishImport(ctx, business, batch.ID)
	if err != nil || finished.Status != "completed" || finished.Created != 3 || finished.Skipped != 2 || finished.Failed != 2 || finished.Updated != 1 {
		t.Fatalf("finished = %+v, %v", finished, err)
	}
	if _, err := svc.ImportLeadRows(ctx, business, actor, batch.ID, domain.ImportOptions{}, rows[:1]); !errors.Is(err, domain.ErrImportClosed) {
		t.Fatalf("rows after finishing: err = %v, want ErrImportClosed", err)
	}

	// Someone works one imported lead; undo keeps it.
	if _, _, err := svc.LogActivity(ctx, business, actor, domain.ActivityInput{LeadID: &cy.ID, Type: domain.ActivityCall, Outcome: domain.CallNoAnswer}); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}
	undone, err := svc.UndoImport(ctx, business, batch.ID)
	if err != nil || undone.Removed != 2 || undone.Kept != 1 {
		t.Fatalf("undo = %+v, %v", undone, err)
	}
	if _, err := svc.GetLead(ctx, business, cy.ID); err != nil {
		t.Fatalf("the worked lead should survive the undo: %v", err)
	}
}

func TestPipelineEditsProtectDeals(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	stages := domain.DefaultStages()
	pipeline, err := svc.CreatePipeline(ctx, business, domain.PipelineInput{Name: unique(t, "Renewals"), Stages: stages})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	if len(pipeline.Stages) != 6 || pipeline.IsDefault {
		t.Fatalf("pipeline = %+v", pipeline)
	}
	deal, err := svc.CreateDeal(ctx, business, actor, domain.DealInput{
		Title: unique(t, "Renewal"), PipelineID: pipeline.ID, StageID: pipeline.Stages[1].ID, Value: 12000, Billing: domain.BillingYearly,
	})
	if err != nil {
		t.Fatalf("CreateDeal: %v", err)
	}

	edit := domain.PipelineInput{Name: pipeline.Name}
	for i, stage := range pipeline.Stages {
		if i == 1 {
			continue
		}
		id := stage.ID
		edit.Stages = append(edit.Stages, domain.StageInput{ID: &id, Name: stage.Name, Probability: stage.Probability, Kind: stage.Kind})
	}
	if _, err := svc.UpdatePipeline(ctx, business, pipeline.ID, edit); !errors.Is(err, domain.ErrStageInUse) {
		t.Fatalf("removing a stage with deals: err = %v, want ErrStageInUse", err)
	}
	if err := svc.DeletePipeline(ctx, business, pipeline.ID); !errors.Is(err, domain.ErrPipelineInUse) {
		t.Fatalf("deleting a pipeline with deals: err = %v, want ErrPipelineInUse", err)
	}

	moved, err := svc.MoveDeal(ctx, business, actor, deal.ID, pipeline.Stages[2].ID)
	if err != nil || moved.StageID != pipeline.Stages[2].ID {
		t.Fatalf("MoveDeal = %+v, %v", moved, err)
	}
	if _, err := svc.MoveDeal(ctx, business, actor, deal.ID, pipeline.Stages[4].ID); !domain.IsValidation(err) {
		t.Fatalf("moving into the won stage: err = %v, want a validation error", err)
	}
	lost, err := svc.LoseDeal(ctx, business, actor, deal.ID, "Went with a competitor")
	if err != nil || lost.Status != domain.DealLost || lost.LostReason != "Went with a competitor" {
		t.Fatalf("LoseDeal = %+v, %v", lost, err)
	}
	stats, err := svc.DealStats(ctx, business, pipeline.ID, nil, today())
	if err != nil || stats.OpenDeals != 0 || stats.WinRate != 0 {
		t.Fatalf("stats after a loss = %+v, %v", stats, err)
	}
	if err := svc.DeleteDeal(ctx, business, deal.ID); err != nil {
		t.Fatalf("DeleteDeal: %v", err)
	}
	if err := svc.DeletePipeline(ctx, business, pipeline.ID); err != nil {
		t.Fatalf("DeletePipeline once empty: %v", err)
	}
}

func ptr[T any](value T) *T { return &value }
