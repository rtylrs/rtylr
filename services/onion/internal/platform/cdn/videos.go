package cdn

import (
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/platform/logger"
)

// videoCapabilities describes what the CDN can do with video today. It is
// returned by GET /videos/info so clients can feature-detect rather than
// hard-coding assumptions. Transcoding and thumbnail extraction are scaffolded
// in the media-engine but not wired through this service yet.
type videoCapabilities struct {
	UploadSupported    bool     `json:"upload_supported"`
	TransformSupported bool     `json:"transform_supported"`
	AcceptedMIMETypes  []string `json:"accepted_mime_types"`
}

// videoInfo handles GET /videos/info. Video uploads are accepted and served
// back verbatim; on-demand video transformation is not implemented yet.
func (c *CDN) videoInfo(w http.ResponseWriter, r *http.Request) {
	info := videoCapabilities{
		UploadSupported:    true,
		TransformSupported: false,
		AcceptedMIMETypes:  []string{"video/mp4", "video/webm", "video/quicktime"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode video info response", logger.Error(err))
	}
}
