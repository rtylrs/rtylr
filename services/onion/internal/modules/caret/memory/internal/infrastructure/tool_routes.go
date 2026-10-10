package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// toolRouteType marks a learned route from a kind of request to the tool
// that served it. Routes are routing hints, so they never appear in the
// memory list or the saved context.
const toolRouteType = "tool_route"

// SaveToolRoute records that requests with these words were served by tool.
func (r *Repository) SaveToolRoute(ctx context.Context, userID int, tool, words string) error {
	sum := sha256.Sum256([]byte(tool + "|" + words))
	_, err := r.pool.ExecContext(ctx, "INSERT INTO user_memory (user_id,memory_type,`key`,value,confidence) VALUES (?,?,?,?,1) ON DUPLICATE KEY UPDATE updated_at=CURRENT_TIMESTAMP",
		userID, toolRouteType, "route_"+hex.EncodeToString(sum[:8]), tool+"\t"+words)
	return err
}

// ToolRoutes returns the user's most recent learned routes as "tool\twords".
func (r *Repository) ToolRoutes(ctx context.Context, userID int) ([]string, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT value FROM user_memory WHERE user_id=? AND memory_type=? ORDER BY updated_at DESC LIMIT 200`, userID, toolRouteType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}
