package http

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"net/http"

	"buf.build/go/protovalidate"
	"github.com/go-playground/validator/v10"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

var (
	validateInstance          = validator.New()
	protoValidator            = protovalidate.GlobalValidator
	errMultipleJSONBodyValues = stdErrors.New("request body must contain a single JSON value")
)

type MiddlewareFunc func(next http.Handler) http.Handler
type HandlerFunc[Req any, Res any] interface {
	Serve(ctx context.Context, req *Req) (Res, error)
}

type ValidatorFunc[Req any] interface {
	Validate(ctx context.Context, req *Req) error
}

func RegisterGET[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodGet, paths, handler, false)
}

func RegisterGETWithValidation[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodGet, paths, handler, true)
}

func RegisterPOST[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodPost, paths, handler, false)
}

func RegisterPOSTWithValidation[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodPost, paths, handler, true)
}

// RegisterPOSTWithValidationAndMiddleware is the per-route
// middleware-aware variant. Use it for endpoints that need an extra
// layer (the main case today is per-route rate limits on
// /auth/login, /auth/register, /recruit/public/apply). Middlewares
// run in order around the handler — index 0 is outermost.
func RegisterPOSTWithValidationAndMiddleware[Req, Res any](
	_ context.Context,
	router LegacyRouter,
	paths []string,
	handler HandlerFunc[Req, Res],
	middlewares ...MiddlewareFunc,
) {
	registerLegacy(router, http.MethodPost, paths, handler, true, middlewares...)
}

func RegisterPATCHWithValidation[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodPatch, paths, handler, true)
}

func RegisterPUTWithValidation[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodPut, paths, handler, true)
}

func RegisterDELETE[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodDelete, paths, handler, false)
}

func RegisterDELETEWithValidation[Req, Res any](_ context.Context, router LegacyRouter, paths []string, handler HandlerFunc[Req, Res]) {
	registerLegacy(router, http.MethodDelete, paths, handler, true)
}

func wrapHandler[Req, Res any](ctx context.Context, handler HandlerFunc[Req, Res], validate bool) http.HandlerFunc {
	validateResponses := responseProtoValidationEnabled()

	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		switch r.Method {
		case "POST", "PUT", "PATCH":
			if !decodeJSONRequestBody(ctx, w, r, &req) {
				return
			}
		case "DELETE":
			// DELETE can ride with either a JSON body (the FE pattern
			// for {"id": N}) or query params. RFC 9110 doesn't forbid
			// a body, just discourages relying on it. Prefer body when
			// present so handlers don't have to special-case the FE
			// shape; fall back to query params otherwise.
			if r.Body != nil && r.ContentLength != 0 {
				if !decodeJSONRequestBody(ctx, w, r, &req) {
					return
				}
			} else {
				if err := parseQueryParams(r, &req); err != nil {
					WriteOneKitError(r.Context(), w, onekiterrors.New(http.StatusBadRequest, err.Error()))
					return
				}
			}
		case "GET":
			if err := parseQueryParams(r, &req); err != nil {
				WriteOneKitError(r.Context(), w, onekiterrors.New(http.StatusBadRequest, err.Error()))
				return
			}
		}

		reqCtx := r.Context()

		if validate {
			if vErr := validateRequest(reqCtx, handler, &req); vErr != nil {
				WriteOneKitError(reqCtx, w, vErr)

				return
			}
		}

		res, serverError := handler.Serve(reqCtx, &req)
		if serverError != nil {
			WriteOneKitError(reqCtx, w, serverError)

			return
		}

		if validateResponses {
			if err := validateProtoMessages(res); err != nil {
				logger.Ctx(reqCtx).Error("Response proto validation failed", logger.Error(err))
				WriteOneKitError(reqCtx, w, onekiterrors.New(http.StatusInternalServerError, "Response validation failed"))

				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Ctx(ctx).Error("Failed to encode response", logger.Error(err))
		}
	}
}
