package api

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/application"
)

func RegisterRoutes(_ context.Context, router serverHttp.LegacyRouter, service *application.Service, permissions PermissionChecker) {
	handler := &GenerateReportHandler{Service: service, Permissions: permissions}
	for _, path := range GenerateReportRoutes() {
		// The handler supports both JSON and CSV responses and therefore owns
		// response content negotiation in addition to request validation.
		router.Post(path, handler.ServeHTTP)
	}
}
