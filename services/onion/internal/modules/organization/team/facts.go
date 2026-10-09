package team

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/team/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// Facts supplies committed team audit facts in local sequence order.
func Facts(db *sql.DB) eventbus.FactSource { return outbox.NewSource(db) }
