package routes

import (
	"context"
	"encoding/json"
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

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Stats
	serverHttp.RegisterGET[handlers.StatsRequest, handlers.StatsResponse](
		ctx, router, []string{"/console/stats"},
		&handlers.StatsHandler{Service: service},
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

func getIDParam(r *http.Request, param string) (int64, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

func checkAdmin(r *http.Request) (*authPkg.AuthUser, bool) {
	user, ok := authPkg.GetAuthUser(r.Context())
	if !ok || user.Role != "console_admin" {
		return nil, false
	}
	return user, true
}

// ── raw handlers ──────────────────────────────────────────────────────────────

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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid body")
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
		_ = json.NewDecoder(r.Body).Decode(&req)
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
		_ = json.NewDecoder(r.Body).Decode(&req)
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid body")
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid body")
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid body")
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
