package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

type MiddlewareFunc func(next http.Handler) http.Handler

func ApplyMiddlewares(r *chi.Mux, middlewares []MiddlewareFunc) {
	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}

type HandlerFunc[Req any, Res any] interface {
	Serve(ctx context.Context, req *Req) (Res, *errors.ServerError)
}

type ValidatorFunc[Req any] interface {
	Validate(ctx context.Context, req *Req) *errors.ServerError
}

func RegisterGET[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Get(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterGETWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Get(path, wrapHandler(ctx, handler, true))
	}
}

func RegisterPOST[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Post(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterPOSTWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Post(path, wrapHandler(ctx, handler, true))
	}
}

// RegisterPOSTWithValidationAndMiddleware is the per-route
// middleware-aware variant. Use it for endpoints that need an extra
// layer (the main case today is per-route rate limits on
// /auth/login, /auth/register, /recruit/public/apply). Middlewares
// run in order around the handler — index 0 is outermost.
func RegisterPOSTWithValidationAndMiddleware[Req, Res any](
	ctx context.Context,
	router *chi.Mux,
	paths []string,
	handler HandlerFunc[Req, Res],
	middlewares ...MiddlewareFunc,
) {
	for _, path := range paths {
		var final http.Handler = wrapHandler(ctx, handler, true)
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		router.Method(http.MethodPost, path, final)
	}
}

func RegisterPUT[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Put(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterPUTWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Put(path, wrapHandler(ctx, handler, true))
	}
}

func RegisterDELETE[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Delete(path, wrapHandler(ctx, handler, false))
	}
}

func RegisterDELETEWithValidation[Req, Res any](ctx context.Context, router *chi.Mux, paths []string, handler HandlerFunc[Req, Res]) {
	for _, path := range paths {
		router.Delete(path, wrapHandler(ctx, handler, true))
	}
}

func wrapHandler[Req, Res any](ctx context.Context, handler HandlerFunc[Req, Res], validate bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req

		switch r.Method {
		case "POST", "PUT", "PATCH":
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
				writeBadRequest(ctx, w)

				return
			}
		case "DELETE":
			// DELETE can ride with either a JSON body (the FE pattern
			// for {"id": N}) or query params. RFC 9110 doesn't forbid
			// a body, just discourages relying on it. Prefer body when
			// present so handlers don't have to special-case the FE
			// shape; fall back to query params otherwise.
			if r.Body != nil && r.ContentLength != 0 {
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
					writeBadRequest(ctx, w)

					return
				}
			} else {
				parseQueryParams(r, &req)
			}
		case "GET":
			parseQueryParams(r, &req)
		}

		reqCtx := r.Context()

		if validate {
			if vErr := validateRequest(reqCtx, handler, &req); vErr != nil {
				writeServerError(reqCtx, w, vErr)

				return
			}
		}

		res, serverError := handler.Serve(reqCtx, &req)
		if serverError != nil {
			writeServerError(reqCtx, w, serverError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Ctx(ctx).Error("Failed to encode response", logger.Error(err))
		}
	}
}

func writeBadRequest(ctx context.Context, w http.ResponseWriter) {
	logger.Ctx(ctx).Warn("Bad request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := errors.ErrorResponse{
		Error: errors.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid request",
		},
		Status: "error",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(err))
	}
}

func validateRequest[Req any](ctx context.Context, handler interface{}, req *Req) *errors.ServerError {
	if v, ok := handler.(ValidatorFunc[Req]); ok {
		if err := v.Validate(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func parseQueryParams[Req any](r *http.Request, req *Req) {
	query := r.URL.Query()
	if len(query) == 0 {
		return
	}

	queryMap := make(map[string]any)

	for key, values := range query {
		val := values[0]
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			queryMap[key] = n
		} else if f, err := strconv.ParseFloat(val, 64); err == nil {
			queryMap[key] = f
		} else if val == "true" || val == "false" {
			queryMap[key] = val == "true"
		} else {
			queryMap[key] = val
		}
	}

	queryJSON, err := json.Marshal(queryMap)
	if err != nil {
		return
	}

	_ = json.Unmarshal(queryJSON, req) //nolint:errcheck // best-effort query param parsing
}

func writeServerError(ctx context.Context, w http.ResponseWriter, serverError *errors.ServerError) {
	w.Header().Set("Content-Type", "application/json")

	// Most call sites set Code (the app-level status) and leave HTTPCode
	// blank. Falling back to Code keeps a 400 from being silently
	// promoted to a 500 just because HTTPCode wasn't filled in. Only
	// promote to 500 when neither is set.
	if serverError.HTTPCode == 0 {
		if serverError.Code >= 400 && serverError.Code < 600 {
			serverError.HTTPCode = serverError.Code
		} else {
			serverError.HTTPCode = http.StatusInternalServerError
		}
	}

	w.WriteHeader(serverError.HTTPCode)

	response := errors.ErrorResponse{
		Error:  *serverError,
		Status: "error",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(err))
	}
}
