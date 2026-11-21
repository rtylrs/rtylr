package application

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

const (
	fieldOutsider      = int64(880003)
	fieldOtherBusiness = int64(5299)
)

func createField(ctx context.Context, t *testing.T, svc *Service, in domain.CustomFieldInput) domain.CustomField {
	t.Helper()
	f, err := svc.CreateProjectField(ctx, v2UserA, v2Business, in)
	if err != nil {
		t.Fatalf("CreateProjectField(%s): %v", in.Name, err)
	}
	return f
}

func fieldTask(ctx context.Context, t *testing.T, svc *Service, projectID int64, title string, values ...domain.FieldValueInput) domain.Task {
	t.Helper()
	task, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: projectID, BusinessID: v2Business, Title: title, ReporterID: v2UserA, OwnerID: v2UserA,
		FieldValues: values,
	})
	if err != nil {
		t.Fatalf("CreateTask(%s): %v", title, err)
	}
	return task
}

func fieldCode(err error) string {
	var fe *domain.FieldValueError
	if errors.As(err, &fe) {
		return fe.Code
	}
	return ""
}

func TestCustomFields_Lifecycle(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "fields", "", v2UserA)

	sev := createField(ctx, t, svc, domain.CustomFieldInput{
		ProjectID: p.ID, Name: "Severity", Type: domain.FieldTypeSelect,
		Options: []domain.FieldOption{{Label: "Low"}, {Label: "High"}},
	})
	points := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: p.ID, Name: "Points", Type: domain.FieldTypeNumber})
	if _, err := svc.CreateProjectField(ctx, v2UserA, v2Business, domain.CustomFieldInput{
		ProjectID: p.ID, Name: "severity", Type: domain.FieldTypeText,
	}); !errors.Is(err, domain.ErrFieldNameTaken) {
		t.Fatalf("duplicate name err = %v", err)
	}

	low, high := sev.Options[0].ID, sev.Options[1].ID
	a := fieldTask(ctx, t, svc, p.ID, "a",
		domain.FieldValueInput{FieldID: sev.ID, Value: domain.FieldValue{Value: low}},
		domain.FieldValueInput{FieldID: points.ID, Value: domain.FieldValue{Value: "3.50"}})
	if a.FieldValues[sev.ID].Value != low || a.FieldValues[points.ID].Value != "3.5" {
		t.Fatalf("created values = %+v", a.FieldValues)
	}
	b := fieldTask(ctx, t, svc, p.ID, "b", domain.FieldValueInput{FieldID: sev.ID, Value: domain.FieldValue{Value: high}})

	fields, err := svc.ListProjectFields(ctx, p.ID, v2UserA, v2Business)
	if err != nil || len(fields) != 2 || fields[0].ID != sev.ID || fields[0].ValueCount != 2 || fields[1].ValueCount != 1 {
		t.Fatalf("ListProjectFields = %+v, %v", fields, err)
	}

	// Rename the field and an option, drop High: values on Low survive, High is cleared.
	updated, err := svc.UpdateProjectField(ctx, sev.ID, v2UserA, v2Business, domain.CustomFieldInput{
		Name: "Impact", Options: []domain.FieldOption{{ID: low, Label: "Minor"}},
	})
	if err != nil || updated.Name != "Impact" || len(updated.Options) != 1 || updated.Type != domain.FieldTypeSelect {
		t.Fatalf("UpdateProjectField = %+v, %v", updated, err)
	}
	detail, err := svc.GetTaskDetail(ctx, a.ID)
	if err != nil || detail.Task.FieldValues[sev.ID].Value != low {
		t.Fatalf("rename lost the value: %+v, %v", detail.Task.FieldValues, err)
	}
	detail, err = svc.GetTaskDetail(ctx, b.ID)
	if err != nil || !detail.Task.FieldValues[sev.ID].IsEmpty() {
		t.Fatalf("removed option still set: %+v, %v", detail.Task.FieldValues, err)
	}

	// Clear one value, then delete the field and get the count back.
	if _, err := svc.SetTaskFieldValues(ctx, a.ID, v2UserA, v2Business, []domain.FieldValueInput{{FieldID: points.ID}}); err != nil {
		t.Fatalf("clear: %v", err)
	}
	_, cleared, err := svc.DeleteProjectField(ctx, sev.ID, v2UserA, v2Business)
	if err != nil || cleared != 1 {
		t.Fatalf("DeleteProjectField cleared %d, %v; want 1", cleared, err)
	}
	if fields, _ := svc.ListProjectFields(ctx, p.ID, v2UserA, v2Business); len(fields) != 1 || fields[0].ID != points.ID {
		t.Fatalf("after delete fields = %+v", fields)
	}
}

func TestCustomFields_RequiredAndDefaultOnCreate(t *testing.T) {
	svc, ctx, db := v2Svc(t)
	p := createV2Project(ctx, t, svc, "fields-required", "", v2UserA)
	team := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: p.ID, Name: "Team", Type: domain.FieldTypeText, Required: true})
	done := createField(ctx, t, svc, domain.CustomFieldInput{
		ProjectID: p.ID, Name: "Reviewed", Type: domain.FieldTypeCheckbox, Default: domain.FieldValue{Value: "false"},
	})

	var before int
	_ = db.QueryRow("SELECT COUNT(*) FROM flow_tasks WHERE business_id = ? AND project_id = ?", v2Business, p.ID).Scan(&before)
	_, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, Title: "x", ReporterID: v2UserA, OwnerID: v2UserA})
	if fieldCode(err) != domain.FieldErrRequired {
		t.Fatalf("missing required err = %v", err)
	}
	var after int
	_ = db.QueryRow("SELECT COUNT(*) FROM flow_tasks WHERE business_id = ? AND project_id = ?", v2Business, p.ID).Scan(&after)
	if after != before {
		t.Fatalf("a refused create left a task behind (%d -> %d)", before, after)
	}

	task := fieldTask(ctx, t, svc, p.ID, "ok", domain.FieldValueInput{FieldID: team.ID, Value: domain.FieldValue{Value: "Ops"}})
	if task.FieldValues[done.ID].Value != "false" {
		t.Fatalf("default not applied: %+v", task.FieldValues)
	}
	if _, err := svc.SetTaskFieldValues(ctx, task.ID, v2UserA, v2Business, []domain.FieldValueInput{{FieldID: team.ID}}); fieldCode(err) != domain.FieldErrRequired {
		t.Fatalf("clearing a required field err = %v", err)
	}
}

func TestCustomFields_ValuesStayInTheirProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "fields-a", "", v2UserA)
	other := createV2Project(ctx, t, svc, "fields-b", "", v2UserA)
	foreign := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: other.ID, Name: "Foreign", Type: domain.FieldTypeText})
	task := fieldTask(ctx, t, svc, p.ID, "local")

	if _, err := svc.SetTaskFieldValues(ctx, task.ID, v2UserA, v2Business, []domain.FieldValueInput{
		{FieldID: foreign.ID, Value: domain.FieldValue{Value: "x"}},
	}); fieldCode(err) != domain.FieldErrUnknownField {
		t.Fatalf("foreign field err = %v", err)
	}
	if _, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: p.ID, BusinessID: v2Business, Title: "y", ReporterID: v2UserA, OwnerID: v2UserA,
		FieldValues: []domain.FieldValueInput{{FieldID: foreign.ID, Value: domain.FieldValue{Value: "x"}}},
	}); fieldCode(err) != domain.FieldErrUnknownField {
		t.Fatalf("create with foreign field err = %v", err)
	}

	people := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: p.ID, Name: "Reviewer", Type: domain.FieldTypePerson})
	if _, err := svc.SetTaskFieldValues(ctx, task.ID, v2UserA, v2Business, []domain.FieldValueInput{
		{FieldID: people.ID, Value: domain.FieldValue{Values: []string{"880003"}}},
	}); fieldCode(err) != domain.FieldErrPerson {
		t.Fatalf("person who cannot open the project err = %v", err)
	}
}

func TestCustomFields_FilterAndSortTasks(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "fields-filter", "", v2UserA)
	sev := createField(ctx, t, svc, domain.CustomFieldInput{
		ProjectID: p.ID, Name: "Severity", Type: domain.FieldTypeSelect,
		Options: []domain.FieldOption{{Label: "Low"}, {Label: "High"}},
	})
	pts := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: p.ID, Name: "Points", Type: domain.FieldTypeNumber})
	tags := createField(ctx, t, svc, domain.CustomFieldInput{
		ProjectID: p.ID, Name: "Tags", Type: domain.FieldTypeMultiSelect,
		Options: []domain.FieldOption{{Label: "ui"}, {Label: "api"}},
	})
	low, high := sev.Options[0].ID, sev.Options[1].ID
	ui, api := tags.Options[0].ID, tags.Options[1].ID
	val := func(f domain.CustomField, v string) domain.FieldValueInput {
		return domain.FieldValueInput{FieldID: f.ID, Value: domain.FieldValue{Value: v}}
	}
	many := func(f domain.CustomField, v ...string) domain.FieldValueInput {
		return domain.FieldValueInput{FieldID: f.ID, Value: domain.FieldValue{Values: v}}
	}
	t1 := fieldTask(ctx, t, svc, p.ID, "t1", val(sev, high), val(pts, "8"), many(tags, ui))
	t2 := fieldTask(ctx, t, svc, p.ID, "t2", val(sev, low), val(pts, "2"), many(tags, api, ui))
	t3 := fieldTask(ctx, t, svc, p.ID, "t3")

	list := func(filter []domain.RawFieldFilter, sortBy, dir string) []int64 {
		t.Helper()
		filters, sort, err := svc.ResolveTaskFieldQuery(ctx, v2UserA, v2Business, p.ID, filter, sortBy, dir)
		if err != nil {
			t.Fatalf("ResolveTaskFieldQuery(%+v): %v", filter, err)
		}
		tasks, _, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{
			ProjectID: p.ID, Limit: 50, FieldFilters: filters, FieldSort: sort,
		})
		if err != nil {
			t.Fatalf("ListTasksV2: %v", err)
		}
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		return ids
	}
	same := func(name string, got []int64, want ...int64) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s = %v, want %v", name, got, want)
			}
		}
	}
	clause := func(f domain.CustomField, op string, values ...string) []domain.RawFieldFilter {
		return []domain.RawFieldFilter{{FieldID: f.ID, Op: op, Values: values}}
	}

	same("select any high", list(clause(sev, "any", high), "", ""), t1.ID)
	same("multi any api", list(clause(tags, "any", api), "", ""), t2.ID)
	same("number range", list(clause(pts, "range", "3", ""), "", ""), t1.ID)
	same("sort points asc, empty last", list(nil, "field:"+itoa(pts.ID), "asc"), t2.ID, t1.ID, t3.ID)
	same("sort select by option order desc", list(nil, "field:"+itoa(sev.ID), "desc"), t1.ID, t2.ID, t3.ID)

	if _, _, err := svc.ResolveTaskFieldQuery(ctx, v2UserA, v2Business, p.ID, nil, "field:"+itoa(tags.ID), "asc"); !errors.Is(err, domain.ErrFieldSortInvalid) {
		t.Fatalf("multi select sort err = %v", err)
	}
	if _, _, err := svc.ResolveTaskFieldQuery(ctx, v2UserA, v2Business, 0, clause(sev, "eq", low), "", ""); !errors.Is(err, domain.ErrFieldFilterInvalid) {
		t.Fatalf("filter without project err = %v", err)
	}
	tasks, _, _ := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{ProjectID: p.ID, Limit: 50})
	for _, task := range tasks {
		if task.ID == t2.ID && len(task.FieldValues[tags.ID].Values) != 2 {
			t.Fatalf("listing did not attach values: %+v", task.FieldValues)
		}
	}
}

// TestCustomFields_AccessMatrix pins who may do what: definitions are read by
// anyone who opens the project and managed by its admins; values follow the
// task's member tier; nothing crosses a business.
func TestCustomFields_AccessMatrix(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "fields-access", "", v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleMember, AddedBy: v2UserA}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}
	field := createField(ctx, t, svc, domain.CustomFieldInput{ProjectID: p.ID, Name: "Note", Type: domain.FieldTypeText})
	task := fieldTask(ctx, t, svc, p.ID, "guarded")
	write := []domain.FieldValueInput{{FieldID: field.ID, Value: domain.FieldValue{Value: "hi"}}}
	rename := domain.CustomFieldInput{Name: "Renamed"}

	type call struct {
		name string
		run  func(user, business int64) error
	}
	calls := []call{
		{"list", func(u, b int64) error { _, err := svc.ListProjectFields(ctx, p.ID, u, b); return err }},
		{"reorder", func(u, b int64) error {
			_, err := svc.ReorderProjectFields(ctx, p.ID, u, b, []int64{field.ID})
			return err
		}},
		{"create", func(u, b int64) error {
			_, err := svc.CreateProjectField(ctx, u, b, domain.CustomFieldInput{ProjectID: p.ID, Name: "New" + itoa(u+b), Type: domain.FieldTypeText})
			return err
		}},
		{"update", func(u, b int64) error { _, err := svc.UpdateProjectField(ctx, field.ID, u, b, rename); return err }},
		{"set values", func(u, b int64) error { _, err := svc.SetTaskFieldValues(ctx, task.ID, u, b, write); return err }},
		{"filter", func(u, b int64) error {
			_, _, err := svc.ResolveTaskFieldQuery(ctx, u, b, p.ID, []domain.RawFieldFilter{{FieldID: field.ID, Op: "eq", Values: []string{"x"}}}, "", "")
			return err
		}},
	}
	// allowed[caller][call]
	cases := []struct {
		who      string
		user     int64
		business int64
		allowed  map[string]bool
	}{
		{"project admin", v2UserA, v2Business, map[string]bool{"list": true, "create": true, "update": true, "reorder": true, "set values": true, "filter": true}},
		{"project member", v2UserB, v2Business, map[string]bool{"list": true, "set values": true, "filter": true}},
		{"outsider", fieldOutsider, v2Business, map[string]bool{}},
		{"admin in another business", v2UserA, fieldOtherBusiness, map[string]bool{}},
	}
	for _, c := range cases {
		for _, cl := range calls {
			err := cl.run(c.user, c.business)
			if c.allowed[cl.name] && err != nil {
				t.Errorf("%s %s: err = %v, want allowed", c.who, cl.name, err)
			}
			if !c.allowed[cl.name] && !errors.Is(err, domain.ErrUnauthorized) && !errors.Is(err, domain.ErrFieldNotFound) {
				t.Errorf("%s %s: err = %v, want refused", c.who, cl.name, err)
			}
		}
	}
	if _, _, err := svc.DeleteProjectField(ctx, field.ID, v2UserB, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("member delete err = %v", err)
	}
	if _, _, err := svc.DeleteProjectField(ctx, field.ID, v2UserA, fieldOtherBusiness); !errors.Is(err, domain.ErrFieldNotFound) {
		t.Errorf("other business delete err = %v", err)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
