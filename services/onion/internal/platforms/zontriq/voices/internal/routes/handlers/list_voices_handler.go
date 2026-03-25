package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	voicesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/voices/v1"
	"github.com/voxire/onion/internal/platform/tts"
)

func ListVoicesRoutes() []string {
	return []string{"/zontriq/voices"}
}

type ListVoicesHandler struct{}

var (
	elevenLabsCache     []*voicesv1.Voice
	elevenLabsCacheTime time.Time
	elevenLabsCacheMu   sync.RWMutex
	cacheTTL            = 30 * time.Minute
)

var coquiVoices = []*voicesv1.Voice{
	{Id: "coqui-male-1", Name: "Atlas", Description: "Deep, commanding presence for documentaries", Provider: "coqui", Gender: "male"},
	{Id: "coqui-male-2", Name: "River", Description: "Warm, conversational everyday tone", Provider: "coqui", Gender: "male"},
	{Id: "coqui-male-3", Name: "Blaze", Description: "High-energy delivery for motivation", Provider: "coqui", Gender: "male"},
	{Id: "coqui-female-1", Name: "Nova", Description: "Clear, authoritative documentary style", Provider: "coqui", Gender: "female"},
	{Id: "coqui-female-2", Name: "Luna", Description: "Smooth, calm delivery for relaxing content", Provider: "coqui", Gender: "female"},
}

func (h *ListVoicesHandler) Serve(ctx context.Context, req *voicesv1.ListVoicesRequest) (voicesv1.ListVoicesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return voicesv1.ListVoicesResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	provider := req.GetProvider()
	if provider == "" {
		provider = "coqui"
	}

	switch provider {
	case "coqui":
		return voicesv1.ListVoicesResponse{Voices: coquiVoices}, nil
	case "elevenlabs":
		voices, err := fetchElevenLabsVoices(ctx)
		if err != nil {
			return voicesv1.ListVoicesResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadGateway,
				Message: "Failed to fetch ElevenLabs voices",
			}
		}

		return voicesv1.ListVoicesResponse{Voices: voices}, nil
	default:
		return voicesv1.ListVoicesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Unknown provider: " + provider,
		}
	}
}

func fetchElevenLabsVoices(ctx context.Context) ([]*voicesv1.Voice, error) {
	elevenLabsCacheMu.RLock()

	if elevenLabsCache != nil && time.Since(elevenLabsCacheTime) < cacheTTL {
		cached := elevenLabsCache

		elevenLabsCacheMu.RUnlock()

		return cached, nil
	}

	elevenLabsCacheMu.RUnlock()

	if !tts.IsConfigured() {
		return nil, tts.ErrNotInitialized
	}

	sdkVoices, err := tts.GetVoices(ctx)
	if err != nil {
		return nil, err
	}

	voices := make([]*voicesv1.Voice, 0, len(sdkVoices))
	for i := range sdkVoices {
		v := &sdkVoices[i]
		voice := &voicesv1.Voice{
			Id:          v.VoiceId,
			Name:        v.Name,
			Description: v.Category,
			Provider:    "elevenlabs",
			PreviewUrl:  v.PreviewUrl,
		}

		if v.Labels != nil {
			voice.Gender = v.Labels["gender"]
			voice.Accent = v.Labels["accent"]
		}

		voices = append(voices, voice)
	}

	elevenLabsCacheMu.Lock()

	elevenLabsCache = voices
	elevenLabsCacheTime = time.Now()

	elevenLabsCacheMu.Unlock()

	return voices, nil
}
