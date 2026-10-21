package application

import "testing"

func TestBindToolInputToPageScopeAddsSelectedIDWithoutOverwritingInput(t *testing.T) {
	pageContext := map[string]any{
		"selected": map[string]any{
			"customer_id": "cus_server_scope",
			"name":        "Ada Secret",
		},
	}

	got := BindToolInputToPageScope("crm.get_customer", map[string]any{}, pageContext)
	if got["customer_id"] != "cus_server_scope" {
		t.Fatalf("bound selected id = %#v, want server-side selected id", got["customer_id"])
	}
	if _, ok := got["name"]; ok {
		t.Fatalf("page record values must not be copied into tool input: %#v", got)
	}

	explicit := BindToolInputToPageScope("crm.get_customer", map[string]any{"customer_id": "cus_explicit"}, pageContext)
	if explicit["customer_id"] != "cus_explicit" {
		t.Fatalf("explicit tool input was overwritten: %#v", explicit)
	}
}

func TestBindToolInputToPageScopeDoesNotBindCreateTargets(t *testing.T) {
	got := BindToolInputToPageScope("crm.create_customer", map[string]any{"name": "Ada"}, map[string]any{
		"selected": map[string]any{"id": "existing_customer"},
	})
	if _, ok := got["id"]; ok {
		t.Fatalf("create input inherited an existing selected record: %#v", got)
	}
}
