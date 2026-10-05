package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type (
	AdminCareerRequest    = adminCareerRequest
	AdminTaskRequest      = adminTaskRequest
	AdminShortlistRequest = adminShortlistRequest
)

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type AdminLoginResponse struct {
	Token string     `json:"token"`
	Admin *adminView `json:"admin"`
}

type AdminLoginValidatedHandler struct {
	Service *application.Service
}

func (h *AdminLoginValidatedHandler) Validate(_ context.Context, req *AdminLoginRequest) error {
	if req == nil || strings.TrimSpace(req.Email) == "" || len(req.Password) < 6 {
		return onekiterrors.New(http.StatusBadRequest, "Email and a password of at least 6 characters are required")
	}
	return nil
}

func (h *AdminLoginValidatedHandler) Serve(ctx context.Context, req *AdminLoginRequest) (AdminLoginResponse, error) {
	token, admin, err := h.Service.AdminLogin(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return AdminLoginResponse{}, onekiterrors.New(http.StatusUnauthorized, "Invalid email or password")
		}
		return AdminLoginResponse{}, onekiterrors.New(http.StatusInternalServerError, "Login failed")
	}
	return AdminLoginResponse{Token: token, Admin: adminViewPointer(admin)}, nil
}

type CreateAdminRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name"`
}

type CreateAdminResponse struct {
	Admin *adminView `json:"admin"`
}

type CreateAdminValidatedHandler struct {
	Service *application.Service
}

func (h *CreateAdminValidatedHandler) Validate(_ context.Context, req *CreateAdminRequest) error {
	if req == nil || strings.TrimSpace(req.Email) == "" || len(req.Password) < 6 {
		return onekiterrors.New(http.StatusBadRequest, "Email and a password of at least 6 characters are required")
	}
	return nil
}

func (h *CreateAdminValidatedHandler) Serve(ctx context.Context, req *CreateAdminRequest) (CreateAdminResponse, error) {
	admin, err := h.Service.CreateAdmin(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrAdminExists) {
			return CreateAdminResponse{}, onekiterrors.New(http.StatusConflict, "An admin with that email already exists")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return CreateAdminResponse{}, onekiterrors.New(http.StatusBadRequest, "Email and a password of at least 6 characters are required")
		}
		return CreateAdminResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create admin")
	}
	return CreateAdminResponse{Admin: adminViewPointer(admin)}, nil
}

type CreateCareerValidatedHandler struct {
	Service *application.Service
}

type CareerMutationResponse struct {
	Career *openingView `json:"career"`
}

func (h *CreateCareerValidatedHandler) Validate(_ context.Context, req *adminCareerRequest) error {
	if req == nil || strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	return nil
}

func (h *CreateCareerValidatedHandler) Serve(ctx context.Context, req *adminCareerRequest) (CareerMutationResponse, error) {
	career, err := h.Service.CreateCareer(ctx, req.careerInput(0))
	if err != nil {
		return CareerMutationResponse{}, careerMutationError(err, "create")
	}
	return CareerMutationResponse{Career: openingViewPointer(career)}, nil
}

type UpdateCareerValidatedHandler struct {
	Service *application.Service
}

func (h *UpdateCareerValidatedHandler) Validate(ctx context.Context, req *adminCareerRequest) error {
	if err := validateRecruitID(ctx, "id", "Invalid career id"); err != nil {
		return err
	}
	if req == nil || strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	return nil
}

func (h *UpdateCareerValidatedHandler) Serve(ctx context.Context, req *adminCareerRequest) (CareerMutationResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid career id")
	if err != nil {
		return CareerMutationResponse{}, err
	}
	career, err := h.Service.UpdateCareer(ctx, req.careerInput(id))
	if err != nil {
		return CareerMutationResponse{}, careerMutationError(err, "update")
	}
	return CareerMutationResponse{Career: openingViewPointer(career)}, nil
}

type SetCareerActiveRequest struct {
	Active bool `json:"active"`
}

type SetCareerActiveValidatedHandler struct {
	Service *application.Service
}

func (h *SetCareerActiveValidatedHandler) Validate(ctx context.Context, _ *SetCareerActiveRequest) error {
	return validateRecruitID(ctx, "id", "Invalid career id")
}

func (h *SetCareerActiveValidatedHandler) Serve(ctx context.Context, req *SetCareerActiveRequest) (CareerMutationResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid career id")
	if err != nil {
		return CareerMutationResponse{}, err
	}
	career, err := h.Service.SetCareerActive(ctx, id, req.Active)
	if err != nil {
		return CareerMutationResponse{}, careerMutationError(err, "update active state for")
	}
	return CareerMutationResponse{Career: openingViewPointer(career)}, nil
}

type RetryCVJobRequest struct{}

type StatusResponse struct {
	Status string `json:"status"`
}

type RetryCVJobValidatedHandler struct {
	Service *application.Service
}

func (h *RetryCVJobValidatedHandler) Validate(ctx context.Context, _ *RetryCVJobRequest) error {
	if err := validateRecruitID(ctx, "id", "Invalid CV job id"); err != nil {
		return err
	}
	if source := strings.TrimSpace(chi.URLParamFromCtx(ctx, "source")); source != "career" && source != "talent" {
		return onekiterrors.New(http.StatusBadRequest, "Invalid CV job")
	}
	return nil
}

func (h *RetryCVJobValidatedHandler) Serve(ctx context.Context, _ *RetryCVJobRequest) (StatusResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid CV job id")
	if err != nil {
		return StatusResponse{}, err
	}
	if err := h.Service.RetryCVProcessing(ctx, chi.URLParamFromCtx(ctx, "source"), id); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return StatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid CV job")
		}
		if errors.Is(err, domain.ErrApplicationNotFound) {
			return StatusResponse{}, onekiterrors.New(http.StatusNotFound, "CV job not found")
		}
		return StatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to retry CV job")
	}
	return StatusResponse{Status: "success"}, nil
}

type CreateTaskValidatedHandler struct {
	Service *application.Service
}

type TaskMutationResponse struct {
	Task *taskView `json:"task"`
}

func (h *CreateTaskValidatedHandler) Validate(_ context.Context, req *adminTaskRequest) error {
	if req == nil || req.ApplicationID <= 0 || strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Application and task title are required")
	}
	return nil
}

func (h *CreateTaskValidatedHandler) Serve(ctx context.Context, req *adminTaskRequest) (TaskMutationResponse, error) {
	in, err := req.taskInput(0)
	if err != nil {
		return TaskMutationResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid due date")
	}
	task, err := h.Service.UpsertTask(ctx, in)
	if err != nil {
		return TaskMutationResponse{}, taskMutationError(err, "create")
	}
	return TaskMutationResponse{Task: taskViewPointer(task)}, nil
}

type UpdateTaskValidatedHandler struct {
	Service *application.Service
}

func (h *UpdateTaskValidatedHandler) Validate(ctx context.Context, _ *adminTaskRequest) error {
	return validateRecruitID(ctx, "id", "Invalid task id")
}

func (h *UpdateTaskValidatedHandler) Serve(ctx context.Context, req *adminTaskRequest) (TaskMutationResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid task id")
	if err != nil {
		return TaskMutationResponse{}, err
	}
	in, err := req.taskInput(id)
	if err != nil {
		return TaskMutationResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid due date")
	}
	task, err := h.Service.UpdateTask(ctx, in)
	if err != nil {
		return TaskMutationResponse{}, taskMutationError(err, "update")
	}
	return TaskMutationResponse{Task: taskViewPointer(task)}, nil
}

type CreateShortlistValidatedHandler struct {
	Service *application.Service
}

type ShortlistMutationResponse struct {
	Shortlist *shortlistView `json:"shortlist"`
}

func (h *CreateShortlistValidatedHandler) Validate(_ context.Context, req *adminShortlistRequest) error {
	if req == nil || req.JobID <= 0 || strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Job and shortlist name are required")
	}
	return nil
}

func (h *CreateShortlistValidatedHandler) Serve(ctx context.Context, req *adminShortlistRequest) (ShortlistMutationResponse, error) {
	shortlist, err := h.Service.CreateShortlist(ctx, domain.ShortlistInput{
		JobID: req.JobID, Name: req.Name, Status: req.Status, Notes: req.Notes,
	})
	if err != nil {
		return ShortlistMutationResponse{}, shortlistMutationError(err)
	}
	return ShortlistMutationResponse{Shortlist: shortlistViewPointer(shortlist)}, nil
}

type AddShortlistItemRequest struct {
	ApplicationID int64  `json:"application_id" validate:"required,gt=0"`
	Rank          int    `json:"rank"`
	Notes         string `json:"notes"`
}

type AddShortlistItemResponse struct {
	Item *shortlistItemView `json:"item"`
}

type AddShortlistItemValidatedHandler struct {
	Service *application.Service
}

func (h *AddShortlistItemValidatedHandler) Validate(ctx context.Context, req *AddShortlistItemRequest) error {
	if err := validateRecruitID(ctx, "id", "Invalid shortlist id"); err != nil {
		return err
	}
	if req == nil || req.ApplicationID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Shortlist and application are required")
	}
	return nil
}

func (h *AddShortlistItemValidatedHandler) Serve(ctx context.Context, req *AddShortlistItemRequest) (AddShortlistItemResponse, error) {
	shortlistID, err := recruitID(ctx, "id", "Invalid shortlist id")
	if err != nil {
		return AddShortlistItemResponse{}, err
	}
	item, err := h.Service.AddShortlistItem(ctx, domain.ShortlistItemInput{
		ShortlistID: shortlistID, ApplicationID: req.ApplicationID, Rank: req.Rank, Notes: req.Notes,
	})
	if err != nil {
		return AddShortlistItemResponse{}, shortlistItemMutationError(err)
	}
	return AddShortlistItemResponse{Item: shortlistItemViewPointer(item)}, nil
}

type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type UpdateApplicationStatusValidatedHandler struct {
	Service *application.Service
}

func (h *UpdateApplicationStatusValidatedHandler) Validate(ctx context.Context, req *UpdateApplicationStatusRequest) error {
	if err := validateRecruitID(ctx, "id", "Invalid application id"); err != nil {
		return err
	}
	if req == nil || strings.TrimSpace(req.Status) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	return nil
}

func (h *UpdateApplicationStatusValidatedHandler) Serve(ctx context.Context, req *UpdateApplicationStatusRequest) (StatusResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid application id")
	if err != nil {
		return StatusResponse{}, err
	}
	if err := h.Service.UpdateApplicationStatus(ctx, id, req.Status); err != nil {
		if errors.Is(err, domain.ErrApplicationNotFound) {
			return StatusResponse{}, onekiterrors.New(http.StatusNotFound, "Application not found")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return StatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Status is required")
		}
		return StatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update status")
	}
	return StatusResponse{Status: "success"}, nil
}

type UpdateApplicationReviewRequest struct {
	Status       string `json:"status"`
	ReviewRating int    `json:"review_rating" validate:"min=0,max=5"`
	ReviewNotes  string `json:"review_notes"`
	NextStep     string `json:"next_step"`
}

type UpdateApplicationReviewValidatedHandler struct {
	Service *application.Service
}

func (h *UpdateApplicationReviewValidatedHandler) Validate(ctx context.Context, _ *UpdateApplicationReviewRequest) error {
	return validateRecruitID(ctx, "id", "Invalid application id")
}

func (h *UpdateApplicationReviewValidatedHandler) Serve(ctx context.Context, req *UpdateApplicationReviewRequest) (StatusResponse, error) {
	id, err := recruitID(ctx, "id", "Invalid application id")
	if err != nil {
		return StatusResponse{}, err
	}
	err = h.Service.UpdateApplicationReview(ctx, id, domain.ApplicationReviewInput{
		Status: req.Status, ReviewRating: req.ReviewRating, ReviewNotes: req.ReviewNotes, NextStep: req.NextStep,
	})
	if err != nil {
		if errors.Is(err, domain.ErrApplicationNotFound) {
			return StatusResponse{}, onekiterrors.New(http.StatusNotFound, "Application not found")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return StatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid review")
		}
		return StatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update review")
	}
	return StatusResponse{Status: "success"}, nil
}

func validateRecruitID(ctx context.Context, param, message string) error {
	if _, err := recruitID(ctx, param, message); err != nil {
		return err
	}
	return nil
}

func recruitID(ctx context.Context, param, message string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParamFromCtx(ctx, param), 10, 64)
	if err != nil || id <= 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, message)
	}
	return id, nil
}

func careerMutationError(err error, action string) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	if errors.Is(err, domain.ErrCareerNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Career not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to "+action+" career")
}

func taskMutationError(err error, action string) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return onekiterrors.New(http.StatusBadRequest, "Application and task title are required")
	}
	if errors.Is(err, domain.ErrApplicationNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Task or application not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to "+action+" task")
}

func shortlistMutationError(err error) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return onekiterrors.New(http.StatusBadRequest, "Job and shortlist name are required")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to create shortlist")
}

func shortlistItemMutationError(err error) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return onekiterrors.New(http.StatusBadRequest, "Shortlist and application are required")
	}
	if errors.Is(err, domain.ErrApplicationNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Shortlist item not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to add shortlist item")
}
