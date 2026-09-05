package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/console/internal/application"
)

func writeSelfHostLicenseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrSelfHostLicenseNotFound):
		writeError(w, http.StatusNotFound, "Self-host license not found")
	case errors.Is(err, application.ErrInvalidSelfHostLicenseInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, application.ErrSelfHostLicenseSigningKeyMissing):
		writeError(w, http.StatusInternalServerError, "License signing key is not configured")
	default:
		writeError(w, http.StatusInternalServerError, "Self-host license request failed")
	}
}

func rawListSelfHostLicenses(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		limit := getPositiveQueryInt(r, "limit", 50)
		offset := getPositiveQueryInt(r, "offset", 0)
		email := r.URL.Query().Get("email")

		result, err := svc.ListSelfHostLicenses(r.Context(), limit, offset, email)
		if err != nil {
			writeSelfHostLicenseError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}
