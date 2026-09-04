package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func ContactRoutes() []string { return []string{"/voxhire/contact"} }

// ContactHandler accepts company hiring inquiries from the
// "Tell us about the role" form on hire.voxire.com.
type ContactHandler struct{ Service *application.Service }

// ContactRequest mirrors the typed ContactPayload in voxhire-web/src/lib/api.ts.
// Field names use the JSON wire shape that the frontend already sends.
type ContactRequest struct {
	CompanyName string `json:"company_name"`
	YourName    string `json:"your_name"`
	YourTitle   string `json:"your_title"`
	WorkEmail   string `json:"work_email"`
	Phone       string `json:"phone"`
	Role        string `json:"role"`
	Headcount   string `json:"headcount"`
	Timeline    string `json:"timeline"`
	Service     string `json:"service"`
	Message     string `json:"message"`
	AgencyOptIn bool   `json:"agency_opt_in"`
}

// UnmarshalJSON accepts the older camelCase form while exposing only the
// canonical snake_case tags on the request type.
func (r *ContactRequest) UnmarshalJSON(data []byte) error {
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decode := func(dst any, names ...string) error {
		for _, name := range names {
			raw, ok := fields[name]
			if !ok {
				continue
			}
			return json.Unmarshal(raw, dst)
		}
		return nil
	}
	var parsed ContactRequest
	if err := decode(&parsed.CompanyName, "company_name", "companyName"); err != nil {
		return err
	}
	if err := decode(&parsed.YourName, "your_name", "yourName"); err != nil {
		return err
	}
	if err := decode(&parsed.YourTitle, "your_title", "yourTitle"); err != nil {
		return err
	}
	if err := decode(&parsed.WorkEmail, "work_email", "workEmail"); err != nil {
		return err
	}
	if err := decode(&parsed.Phone, "phone"); err != nil {
		return err
	}
	if err := decode(&parsed.Role, "role"); err != nil {
		return err
	}
	if err := decode(&parsed.Headcount, "headcount"); err != nil {
		return err
	}
	if err := decode(&parsed.Timeline, "timeline"); err != nil {
		return err
	}
	if err := decode(&parsed.Service, "service"); err != nil {
		return err
	}
	if err := decode(&parsed.Message, "message"); err != nil {
		return err
	}
	if err := decode(&parsed.AgencyOptIn, "agency_opt_in", "agencyOptIn"); err != nil {
		return err
	}
	*r = parsed
	return nil
}

type ContactResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (h *ContactHandler) Validate(_ context.Context, req *ContactRequest) error {
	if req.CompanyName == "" {
		return onekiterrors.New(http.StatusBadRequest, "company_name is required")
	}
	if req.YourName == "" {
		return onekiterrors.New(http.StatusBadRequest, "your_name is required")
	}
	if req.WorkEmail == "" {
		return onekiterrors.New(http.StatusBadRequest, "work_email is required")
	}
	if req.Role == "" {
		return onekiterrors.New(http.StatusBadRequest, "role is required")
	}
	return nil
}

func (h *ContactHandler) Serve(ctx context.Context, req *ContactRequest) (ContactResponse, error) {
	id, err := h.Service.SubmitCompanyLead(ctx, domain.CompanyLeadInput{
		CompanyName:  req.CompanyName,
		ContactName:  req.YourName,
		ContactTitle: req.YourTitle,
		WorkEmail:    req.WorkEmail,
		Phone:        req.Phone,
		Role:         req.Role,
		Headcount:    req.Headcount,
		Timeline:     req.Timeline,
		Service:      req.Service,
		Message:      req.Message,
		AgencyOptIn:  req.AgencyOptIn,
	})
	if err != nil {
		return ContactResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit hiring inquiry")
	}

	return ContactResponse{
		ID:      id,
		Message: "Message received",
		Status:  "success",
	}, nil
}
