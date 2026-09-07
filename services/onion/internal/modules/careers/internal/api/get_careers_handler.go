package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func CareersRoutes() []string {
	return []string{
		"/careers",
	}
}

type CareersHandler struct {
	CareersService *application.Service
}

type GetActiveCareersRequest struct{}

type Career struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Department     string `json:"department"`
	Location       string `json:"location"`
	EmploymentType string `json:"employment_type"`
	WorkMode       string `json:"work_mode"`
	Seniority      string `json:"seniority"`
	Details        string `json:"details"`
	IsActive       bool   `json:"is_active"`
}

type GetActiveCareersResponse struct {
	Careers []*Career `json:"careers"`
}

func newCareer(career domain.Career) *Career {
	return &Career{
		ID:             career.Id,
		Title:          career.Title,
		Department:     career.Department,
		Location:       career.Location,
		EmploymentType: career.EmploymentType,
		WorkMode:       career.WorkMode,
		Seniority:      career.Seniority,
		Details:        career.Details,
		IsActive:       career.IsActive,
	}
}

func (h *CareersHandler) Serve(ctx context.Context, _ *GetActiveCareersRequest) (GetActiveCareersResponse, error) {
	domainCareers, err := h.CareersService.GetActiveCareers(ctx)
	if err != nil {
		return GetActiveCareersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get active careers")
	}

	careers := make([]*Career, 0, len(domainCareers))
	for i := range domainCareers {
		careers = append(careers, newCareer(domainCareers[i]))
	}

	return GetActiveCareersResponse{
		Careers: careers,
	}, nil
}
