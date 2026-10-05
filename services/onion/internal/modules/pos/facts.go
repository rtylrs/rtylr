package pos

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// Facts supplies committed facts in local sequence order. Acknowledgement is
// separate from downstream commits; failed acknowledgement safely retries.
func Facts(db *sql.DB) eventbus.FactSource { return outbox.NewSource(db) }
