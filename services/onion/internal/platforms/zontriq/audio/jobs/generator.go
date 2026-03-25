package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	rpEndpoint "github.com/runpod/go-sdk/pkg/sdk/endpoint"
	"github.com/voxire/onion/internal/platform/gpu"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
	"github.com/voxire/onion/internal/platform/tts"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/domain"
)

var ErrUnsupportedTTS = errors.New("unsupported TTS provider")

type AudioGenerator struct {
	audioService *application.Service
	pollInterval time.Duration
}

func NewAudioGenerator(audioService *application.Service) *AudioGenerator {
	return &AudioGenerator{
		audioService: audioService,
		pollInterval: 5 * time.Second,
	}
}

func (g *AudioGenerator) Start(ctx context.Context) {
	logger.Ctx(ctx).Info("Audio job runner started")

	ticker := time.NewTicker(g.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Ctx(ctx).Info("Audio job runner stopped")

			return
		case <-ticker.C:
			g.processQueuedAudio(ctx)
			g.checkGeneratingAudio(ctx)
		}
	}
}

func (g *AudioGenerator) processQueuedAudio(ctx context.Context) {
	clips, err := g.audioService.GetQueuedAudio(ctx)
	if err != nil {
		return
	}

	for i := range clips {
		clip := &clips[i]

		logger.Ctx(ctx).Info("Processing audio clip",
			logger.Any("clip_id", clip.ID),
			logger.String("provider", clip.Provider),
		)

		switch clip.Provider {
		case "elevenlabs":
			g.processElevenLabs(ctx, clip)
		case "coqui":
			g.submitCoquiToRunPod(ctx, clip)
		default:
			if stateErr := g.audioService.UpdateAudioState(ctx, clip.ID, "failed", "", 0, fmt.Sprintf("unsupported provider: %s", clip.Provider)); stateErr != nil {
				logger.Ctx(ctx).Error("Failed to update state", logger.Error(stateErr))
			}
		}
	}
}

func (g *AudioGenerator) processElevenLabs(ctx context.Context, clip *domain.AudioClip) {
	if updateErr := g.audioService.UpdateAudioState(ctx, clip.ID, "generating", "", 0, ""); updateErr != nil {
		logger.Ctx(ctx).Error("Failed to update state to generating", logger.Error(updateErr))

		return
	}

	audioData, err := tts.TextToSpeech(ctx, clip.VoiceID, clip.Script)
	if err != nil {
		logger.Ctx(ctx).Error("ElevenLabs TTS failed", logger.Any("clip_id", clip.ID), logger.Error(err))
		g.handleFailure(ctx, clip, err)

		return
	}

	g.uploadAndFinish(ctx, clip, audioData)
}

func (g *AudioGenerator) submitCoquiToRunPod(ctx context.Context, clip *domain.AudioClip) {
	if !gpu.IsConfigured() {
		g.handleFailure(ctx, clip, fmt.Errorf("RunPod not configured"))

		return
	}

	input := map[string]any{
		"action":   "tts",
		"text":     clip.Script,
		"voice_id": clip.VoiceID,
		"language": "en",
	}

	jobID, err := gpu.RunAsync(ctx, input)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to submit Coqui job to RunPod", logger.Error(err))
		g.handleFailure(ctx, clip, err)

		return
	}

	if setErr := g.audioService.SetExternalJobID(ctx, clip.ID, jobID); setErr != nil {
		logger.Ctx(ctx).Error("Failed to set external job ID", logger.Error(setErr))

		return
	}

	logger.Ctx(ctx).Info("Coqui TTS job submitted to RunPod",
		logger.Any("clip_id", clip.ID),
		logger.String("job_id", jobID),
	)
}

func (g *AudioGenerator) checkGeneratingAudio(ctx context.Context) {
	clips, err := g.audioService.GetGeneratingAudio(ctx)
	if err != nil {
		return
	}

	for i := range clips {
		clip := &clips[i]
		if clip.ExternalJobID == "" {
			continue
		}

		status, err := gpu.GetStatus(ctx, clip.ExternalJobID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to check RunPod job status", logger.Any("clip_id", clip.ID), logger.Error(err))

			continue
		}

		if status.Status == nil {
			continue
		}

		switch *status.Status {
		case "COMPLETED":
			g.handleRunPodComplete(ctx, clip, status)
		case "FAILED":
			errMsg := "RunPod job failed"
			if status.Error != nil {
				errMsg = *status.Error
			}
			logger.Ctx(ctx).Error("RunPod job failed", logger.Any("clip_id", clip.ID), logger.String("error", errMsg))

			ref := &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}
			g.handleFailure(ctx, ref, fmt.Errorf("%s", errMsg))
		case "IN_QUEUE", "IN_PROGRESS":
			// still processing, check next poll
		}
	}
}

func (g *AudioGenerator) handleRunPodComplete(ctx context.Context, clip *domain.AudioClip, status *rpEndpoint.StatusOutput) {
	if status.Output == nil {
		g.handleFailure(ctx, &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}, fmt.Errorf("no output from RunPod"))

		return
	}

	outputMap, ok := (*status.Output).(map[string]any)
	if !ok {
		g.handleFailure(ctx, &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}, fmt.Errorf("unexpected RunPod output format"))

		return
	}

	audioURL, ok := outputMap["audio_url"].(string)
	if !ok || audioURL == "" {
		g.handleFailure(ctx, &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}, fmt.Errorf("no audio_url in RunPod response"))

		return
	}

	resp, err := http.Get(audioURL) //nolint:gosec
	if err != nil {
		g.handleFailure(ctx, &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}, err)

		return
	}
	defer resp.Body.Close()

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		g.handleFailure(ctx, &domain.AudioClip{ID: clip.ID, RetryCount: clip.RetryCount}, err)

		return
	}

	g.uploadAndFinish(ctx, &domain.AudioClip{ID: clip.ID}, audioData)
}

func (g *AudioGenerator) uploadAndFinish(ctx context.Context, clip *domain.AudioClip, audioData []byte) {
	var audioURL string

	if storage.IsConfigured() {
		key := storage.GenerateKey("audio", fmt.Sprintf("%d.mp3", clip.ID))

		var uploadErr error

		audioURL, uploadErr = storage.Upload(ctx, key, audioData, "audio/mpeg")
		if uploadErr != nil {
			logger.Ctx(ctx).Error("Failed to upload audio to R2", logger.Error(uploadErr))

			if stateErr := g.audioService.UpdateAudioState(ctx, clip.ID, "failed", "", 0, "upload failed"); stateErr != nil {
				logger.Ctx(ctx).Error("Failed to update state", logger.Error(stateErr))
			}

			return
		}
	} else {
		audioURL = fmt.Sprintf("local://audio/%d.mp3", clip.ID)
	}

	if updateErr := g.audioService.UpdateAudioState(ctx, clip.ID, "ready", audioURL, 0, ""); updateErr != nil {
		logger.Ctx(ctx).Error("Failed to update state to ready", logger.Error(updateErr))
	}

	logger.Ctx(ctx).Info("Audio clip generated",
		logger.Any("clip_id", clip.ID),
		logger.String("url", audioURL),
	)
}

func (g *AudioGenerator) handleFailure(ctx context.Context, clip *domain.AudioClip, err error) {
	if clip.RetryCount < 2 {
		if retryErr := g.audioService.IncrementRetryCount(ctx, clip.ID, err.Error()); retryErr != nil {
			logger.Ctx(ctx).Error("Failed to increment retry count", logger.Error(retryErr))
		}
	} else {
		if stateErr := g.audioService.UpdateAudioState(ctx, clip.ID, "failed", "", 0, err.Error()); stateErr != nil {
			logger.Ctx(ctx).Error("Failed to update state to failed", logger.Error(stateErr))
		}
	}
}
