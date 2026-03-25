package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	careersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/voxire/careers/v1"
	"github.com/voxire/onion/internal/platform/utils/converter"
	"github.com/voxire/onion/internal/platforms/voxire/careers/internal/application"
)

func CareersRoutes() []string {
	return []string{
		"/careers",
	}
}

type CareersHandler struct {
	CareersService *application.Service
}

func (h *CareersHandler) Serve(ctx context.Context, _ *careersv1.GetActiveCareersRequest) (careersv1.GetActiveCareersResponse, *errorModels.ServerError) {
	domainCareers, err := h.CareersService.GetActiveCareers(ctx)
	if err != nil {
		return careersv1.GetActiveCareersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get active careers",
		}
	}

	return careersv1.GetActiveCareersResponse{
		Careers: converter.ConvertSlice[*careersv1.Career](domainCareers),
	}, nil
}
