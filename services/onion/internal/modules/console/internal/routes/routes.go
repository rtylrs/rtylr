package routes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	"github.com/voxire/onion/internal/modules/console/internal/routes/handlers"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

const consoleJSONBodyLimit = 4 << 20

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Generic admin resources
	router.Get("/console/resources", rawListResources(ctx, service))
	router.Get("/console/resources/{table}", rawDescribeResource(ctx, service))
	router.Get("/console/resources/{table}/rows", rawListResourceRows(ctx, service))
	router.Post("/console/resources/{table}/rows", rawCreateResourceRow(ctx, service))
	router.Get("/console/resources/{table}/rows/{id}", rawGetResourceRow(ctx, service))
	router.Put("/console/resources/{table}/rows/{id}", rawUpdateResourceRow(ctx, service))
	router.Delete("/console/resources/{table}/rows/{id}", rawDeleteResourceRow(ctx, service))

	// Stats
	serverHttp.RegisterGET[handlers.StatsRequest, handlers.StatsResponse](
		ctx, router, []string{"/console/stats"},
		&handlers.StatsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.AICacheAnalyticsRequest, handlers.AICacheAnalyticsResponse](
		ctx, router, []string{"/console/ai-cache/analytics"},
		&handlers.AICacheAnalyticsHandler{},
	)

	// Users
	serverHttp.RegisterGET[handlers.ListUsersRequest, handlers.ListUsersResponse](
		ctx, router, []string{"/console/users"},
		&handlers.ListUsersHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateUserRequest, handlers.CreateUserResponse](
		ctx, router, []string{"/console/users"},
		&handlers.CreateUserHandler{Service: service},
	)

	// Users by ID (raw handlers for path params)
	router.Get("/console/users/{id}", rawGetUser(ctx, service))
	router.Put("/console/users/{id}", rawUpdateUser(ctx, service))
	router.Delete("/console/users/{id}", rawDeleteUser(ctx, service))
	router.Post("/console/users/{id}/ban", rawBanUser(ctx, service))
	router.Post("/console/users/{id}/unban", rawUnbanUser(ctx, service))
	router.Post("/console/users/{id}/set-console-admin", rawSetConsoleAdmin(ctx, service))
	router.Get("/console/users/{id}/subscription", rawGetUserSubscription(ctx, service))

	// Businesses
	serverHttp.RegisterGET[handlers.ListBusinessesRequest, handlers.ListBusinessesResponse](
		ctx, router, []string{"/console/businesses"},
		&handlers.ListBusinessesHandler{Service: service},
	)
	router.Get("/console/businesses/{id}", rawGetBusiness(ctx, service))
	router.Put("/console/businesses/{id}", rawUpdateBusiness(ctx, service))
	router.Delete("/console/businesses/{id}", rawDeleteBusiness(ctx, service))

	// Subscriptions
	serverHttp.RegisterGET[handlers.ListSubscriptionsRequest, handlers.ListSubscriptionsResponse](
		ctx, router, []string{"/console/subscriptions"},
		&handlers.ListSubscriptionsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.UpsertSubscriptionRequest, handlers.UpsertSubscriptionResponse](
		ctx, router, []string{"/console/subscriptions/upsert"},
		&handlers.UpsertSubscriptionHandler{Service: service},
	)
	router.Get("/console/subscriptions/{id}", rawGetSubscription(ctx, service))
	router.Put("/console/subscriptions/{id}", rawUpdateSubscription(ctx, service))
	router.Delete("/console/subscriptions/{id}", rawDeleteSubscription(ctx, service))

	// Plans
	serverHttp.RegisterGET[handlers.ListPlansRequest, handlers.ListPlansResponse](
		ctx, router, []string{"/console/plans"},
		&handlers.ListPlansHandler{Service: service},
	)
	router.Put("/console/plans/{id}", rawUpdatePlan(ctx, service))

	// Grants
	serverHttp.RegisterGET[handlers.ListGrantsRequest, handlers.ListGrantsResponse](
		ctx, router, []string{"/console/grants"},
		&handlers.ListGrantsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateGrantRequest, handlers.CreateGrantResponse](
		ctx, router, []string{"/console/grants"},
		&handlers.CreateGrantHandler{Service: service},
	)
	router.Post("/console/grants/{id}/revoke", rawRevokeGrant(ctx, service))
	router.Delete("/console/grants/{id}", rawDeleteGrant(ctx, service))
}

// ── helper ────────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Global().Error("Failed to encode JSON response", logger.Error(err))
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorModels.ErrorResponse{
		Error:  errorModels.ServerError{Code: status, Message: msg},
		Status: "error",
	}); err != nil {
		logger.Global().Error("Failed to encode error response", logger.Error(err))
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, consoleJSONBodyLimit)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid body")
		return false
	}

	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "Invalid body")
		return false
	}

	return true
}

func decodeOptionalJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil || r.Body == http.NoBody || r.ContentLength == 0 {
		return true
	}
	return decodeJSONBody(w, r, dst)
}

func getIDParam(r *http.Request, param string) (int64, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

func getPositiveQueryInt(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func checkAdmin(r *http.Request) (*authPkg.AuthUser, bool) {
	user, ok := authPkg.GetAuthUser(r.Context())
	if !ok || user.Role != "console_admin" {
		return nil, false
	}
	return user, true
}

// ── raw handlers ──────────────────────────────────────────────────────────────

func rawListResources(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		result, err := svc.ListResources(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list resources")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawDescribeResource(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		table := chi.URLParam(r, "table")
		result, err := svc.DescribeResource(r.Context(), table)
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawListResourceRows(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		table := chi.URLParam(r, "table")
		limit := getPositiveQueryInt(r, "limit", 50)
		offset := getPositiveQueryInt(r, "offset", 0)
		result, err := svc.ListResourceRows(r.Context(), table, limit, offset)
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawGetResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		result, err := svc.GetResourceRow(r.Context(), chi.URLParam(r, "table"), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource row not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawCreateResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		var req struct {
			Values map[string]any `json:"values"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if req.Values == nil {
			writeError(w, http.StatusBadRequest, "values are required")
			return
		}
		id, err := svc.CreateResourceRow(r.Context(), chi.URLParam(r, "table"), req.Values)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "message": "Resource row created"})
	}
}

func rawUpdateResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		var req struct {
			Values map[string]any `json:"values"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if req.Values == nil {
			writeError(w, http.StatusBadRequest, "values are required")
			return
		}
		if err := svc.UpdateResourceRow(r.Context(), chi.URLParam(r, "table"), chi.URLParam(r, "id"), req.Values); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Resource row updated"})
	}
}

func rawDeleteResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		if err := svc.DeleteResourceRow(r.Context(), chi.URLParam(r, "table"), chi.URLParam(r, "id")); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Resource row deleted"})
	}
}

func rawGetUser(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		user, err := svc.GetUser(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}

func rawUpdateUser(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			FullName string `json:"full_name"`
			Email    string `json:"email"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if err := svc.UpdateUser(r.Context(), id, req.FullName, req.Email); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "User updated"})
	}
}

func rawDeleteUser(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteUser(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted"})
	}
}

func rawBanUser(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			Reason string `json:"reason"`
		}
		if !decodeOptionalJSONBody(w, r, &req) {
			return
		}
		if err := svc.BanUser(r.Context(), id, req.Reason); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to ban user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "User banned"})
	}
}

func rawUnbanUser(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.UnbanUser(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to unban user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "User unbanned"})
	}
}

func rawSetConsoleAdmin(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			IsConsoleAdmin bool `json:"is_console_admin"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if err := svc.SetConsoleAdmin(r.Context(), id, req.IsConsoleAdmin); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update admin status")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Updated"})
	}
}

func rawGetUserSubscription(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		sub, err := svc.GetSubscriptionByOwner(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "No subscription found")
			return
		}
		writeJSON(w, http.StatusOK, sub)
	}
}

func rawGetBusiness(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		b, err := svc.GetBusiness(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Business not found")
			return
		}
		writeJSON(w, http.StatusOK, b)
	}
}

func rawUpdateBusiness(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if err := svc.UpdateBusiness(r.Context(), id, req.Name); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Business updated"})
	}
}

func rawDeleteBusiness(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteBusiness(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Business deleted"})
	}
}

func rawGetSubscription(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		sub, err := svc.GetSubscription(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Subscription not found")
			return
		}
		writeJSON(w, http.StatusOK, sub)
	}
}

func rawUpdateSubscription(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			PlanID    int64  `json:"plan_id"`
			Status    string `json:"status"`
			SeatCount int    `json:"seat_count"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if err := svc.UpdateSubscription(r.Context(), id, req.PlanID, req.Status, req.SeatCount); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update subscription")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Subscription updated"})
	}
}

func rawDeleteSubscription(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteSubscription(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete subscription")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Subscription deleted"})
	}
}

func rawUpdatePlan(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		var req struct {
			Name                    string `json:"name"`
			PricePerSeatCents       int    `json:"price_per_seat_cents"`
			AnnualPricePerSeatCents int    `json:"annual_price_per_seat_cents"`
			IsActive                bool   `json:"is_active"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if err := svc.UpdatePlan(r.Context(), id, req.Name, req.PricePerSeatCents, req.AnnualPricePerSeatCents, req.IsActive); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update plan")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Plan updated"})
	}
}

func rawRevokeGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.RevokeGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to revoke grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Grant revoked"})
	}
}

func rawDeleteGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Grant deleted"})
	}
}
