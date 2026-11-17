package workspaces

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// Facts supplies committed Flow audit facts in local sequence order.
func Facts(db *sql.DB) eventbus.FactSource { return outbox.NewSource(db) }
