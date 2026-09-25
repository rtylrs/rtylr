package auth

import "context"

type TokenValidatorResolver interface {
	ResolveTokenValidator(ctx context.Context, productPlatform ProductPlatform) TokenValidator
}

type StaticTokenValidatorResolver struct {
	defaultValidator   TokenValidator
	platformValidators map[ProductPlatform]TokenValidator
}

func NewStaticTokenValidatorResolver(defaultValidator TokenValidator, platformValidators map[ProductPlatform]TokenValidator) *StaticTokenValidatorResolver {
	if platformValidators == nil {
		platformValidators = make(map[ProductPlatform]TokenValidator)
	}

	return &StaticTokenValidatorResolver{
		defaultValidator:   defaultValidator,
		platformValidators: platformValidators,
	}
}

func (r *StaticTokenValidatorResolver) ResolveTokenValidator(_ context.Context, productPlatform ProductPlatform) TokenValidator {
	if r == nil {
		return nil
	}

	if validator, ok := r.platformValidators[productPlatform]; ok && validator != nil {
		return validator
	}

	return r.defaultValidator
}
