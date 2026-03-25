package translation

import (
	"context"

	"github.com/voxire/onion/internal/modules/translation/internal/application"
	"github.com/voxire/onion/internal/modules/translation/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Language string

const (
	English Language = "en"
	Arabic  Language = "ar"
	French  Language = "fr"
)

type Input struct {
	Key string `json:"key"`
	En  string `json:"en"`
	Ar  string `json:"ar"`
	Fr  string `json:"fr"`
}

type Translation struct {
	Key string `json:"key"`
	En  string `json:"en"`
	Ar  string `json:"ar"`
	Fr  string `json:"fr"`
}

type Service struct {
	service *application.Service
}

type ITranslationService interface {
	GetTranslation(ctx context.Context, key string) Translation
	CreateOrUpdateTranslation(ctx context.Context, input Input) (string, bool)
}

var instance *Service

func Initialize(ctx context.Context) {
	appService := application.NewService(ctx)
	instance = &Service{
		service: appService,
	}
}

func (service *Service) GetTranslation(ctx context.Context, key string) Translation {
	translation, success := service.service.GetTranslation(ctx, key)
	if !success {
		return Translation{}
	}

	return Translation{
		Key: translation.KeyText,
		En:  translation.En,
		Ar:  translation.Ar,
		Fr:  translation.Fr,
	}
}

func (service *Service) CreateOrUpdateTranslation(ctx context.Context, input Input) (string, bool) {
	if input.Key == "" || input.En == "" {
		return "", false
	}

	domainInput := domain.TranslationInput{
		KeyText:  input.Key,
		En:       input.En,
		Ar:       input.Ar,
		Fr:       input.Fr,
		KeyGroup: "",
	}

	key, success := service.service.CreateOrUpdateTranslation(ctx, domainInput)
	if !success {
		logger.Ctx(ctx).Error("failed to create translation")

		return "", false
	}

	return key, true
}

func resolveTranslation(language Language, loadedTranslation Translation) string {
	switch language {
	case English:
		return loadedTranslation.En
	case Arabic:
		return loadedTranslation.Ar
	case French:
		return loadedTranslation.Fr
	default:
		return loadedTranslation.Key
	}
}

func T(ctx context.Context, key string, language Language) string {
	return TDefault(ctx, key, language, key)
}

func TDefault(ctx context.Context, key string, language Language, fallback string) string {
	if key == "" {
		return fallback
	}

	if instance == nil {
		return fallback
	}

	translation := instance.GetTranslation(ctx, key)
	if !translation.Valid() {
		return fallback
	}

	return resolveTranslation(language, translation)
}

func CreateOrUpdateTranslation(ctx context.Context, input Input) (string, bool) {
	if instance == nil {
		return "", false
	}

	return instance.CreateOrUpdateTranslation(ctx, input)
}

func (t Translation) WithFallback(fallback string) Translation {
	if t.Key == "" {
		return Translation{
			Key: "",
			En:  fallback,
			Fr:  fallback,
			Ar:  fallback,
		}
	}

	return Translation{
		Key: t.Key,
		En:  t.En,
		Fr:  t.Fr,
		Ar:  t.Ar,
	}
}

func (t Translation) Valid() bool {
	return t.En != ""
}
