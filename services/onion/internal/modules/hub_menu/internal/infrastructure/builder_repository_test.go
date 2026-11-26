package infrastructure

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/infrastructure/persistence/sqlc"
)

func TestMenuProfileFromRow_HidesPlaceholderPublishedJSONWhenUnpublished(t *testing.T) {
	// MySQL returns "{}" from COALESCE(published_config, JSON_OBJECT()) when
	// published_config is NULL on an unpublished draft row.
	unpublished := menuProfileFromRow(sqlc.GetMenuProfileRow{
		ID:              1,
		BusinessID:      42,
		LocationID:      0,
		DraftConfig:     json.RawMessage(`{"mode":"light"}`),
		PublishedConfig: json.RawMessage(`{}`),
		EnabledLocales:  json.RawMessage(`["en","ar"]`),
		IsPublished:     false,
	})
	if unpublished.PublishedConfigJSON != "" {
		t.Fatalf("expected empty PublishedConfigJSON when unpublished, got %q", unpublished.PublishedConfigJSON)
	}
	if unpublished.DraftConfigJSON != `{"mode":"light"}` {
		t.Fatalf("unexpected DraftConfigJSON: %q", unpublished.DraftConfigJSON)
	}

	published := menuProfileFromRow(sqlc.GetMenuProfileRow{
		ID:              1,
		BusinessID:      42,
		LocationID:      0,
		DraftConfig:     json.RawMessage(`{"mode":"light"}`),
		PublishedConfig: json.RawMessage(`{"mode":"dark"}`),
		EnabledLocales:  json.RawMessage(`["en"]`),
		IsPublished:     true,
	})
	if published.PublishedConfigJSON != `{"mode":"dark"}` {
		t.Fatalf("expected published config preserved, got %q", published.PublishedConfigJSON)
	}
}
