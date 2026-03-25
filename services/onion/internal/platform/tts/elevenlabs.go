package tts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/haguro/elevenlabs-go"
	appConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	defaultClient     *elevenlabs.Client
	ErrNotInitialized = errors.New("ElevenLabs TTS not initialized")
	ErrGenFailed      = errors.New("TTS generation failed")
	ErrNoAPIKey       = errors.New("ElevenLabs API key not configured")
)

const (
	DefaultModel   = "eleven_turbo_v2_5"
	MultilingualV2 = "eleven_multilingual_v2"
	EnglishV2      = "eleven_turbo_v2_5"

	defaultTimeout = 120 * time.Second
)

func Init(ctx context.Context) error {
	apiKey := appConfig.GetElevenLabsAPIKey()
	if apiKey == "" {
		logger.Ctx(ctx).Warn("ElevenLabs API key not configured, TTS features disabled")

		return nil
	}

	defaultClient = elevenlabs.NewClient(ctx, apiKey, defaultTimeout)

	logger.Ctx(ctx).Info("ElevenLabs TTS initialized")

	return nil
}

func IsConfigured() bool {
	return defaultClient != nil
}

func TextToSpeech(ctx context.Context, voiceID, text string, opts ...Option) ([]byte, error) {
	if defaultClient == nil {
		return nil, ErrNotInitialized
	}

	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	req := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: cfg.Model,
		VoiceSettings: &elevenlabs.VoiceSettings{
			Stability:       float32(cfg.Stability),
			SimilarityBoost: float32(cfg.SimilarityBoost),
		},
	}

	audioData, err := defaultClient.TextToSpeech(voiceID, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGenFailed, err)
	}

	data, readErr := io.ReadAll(bytes.NewReader(audioData))
	if readErr != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", readErr)
	}

	return data, nil
}

func GetVoices(ctx context.Context) ([]elevenlabs.Voice, error) {
	if defaultClient == nil {
		return nil, ErrNotInitialized
	}

	voices, err := defaultClient.GetVoices()
	if err != nil {
		return nil, fmt.Errorf("failed to get voices: %w", err)
	}

	return voices, nil
}

type Options struct {
	Model           string
	Stability       float64
	SimilarityBoost float64
}

type Option func(*Options)

func defaultOptions() Options {
	return Options{
		Model:           DefaultModel,
		Stability:       0.5,
		SimilarityBoost: 0.75,
	}
}

func WithModel(model string) Option {
	return func(o *Options) { o.Model = model }
}

func WithStability(s float64) Option {
	return func(o *Options) { o.Stability = s }
}

func WithSimilarityBoost(s float64) Option {
	return func(o *Options) { o.SimilarityBoost = s }
}
