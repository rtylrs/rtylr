package cdn

import (
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/platform/logger"
)

// videoInfo is the placeholder for GET /videos/info. Video processing
// (FFmpeg-based thumbnail extraction and transcoding via a media-engine
// subcommand) is not wired yet; the endpoint exists so clients can probe.
func (c *CDN) videoInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "video processing coming soon",
	}); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode video info response", logger.Error(err))
	}
}
