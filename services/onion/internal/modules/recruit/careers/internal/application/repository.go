package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
)

type repository interface {
	GetActiveCareers(context.Context) ([]domain.Career, error)
	SaveApplication(context.Context, domain.ApplicationInput) (int64, error)
	ListFormFields(context.Context, int64, int64) ([]domain.FormField, error)
	ListFormFieldsByCareer(context.Context, int64) ([]domain.FormField, error)
	CreateFormField(context.Context, domain.FormFieldInput) (int64, error)
	UpdateFormField(context.Context, domain.FormFieldInput) error
	DeleteFormField(context.Context, int64, int64) error
	GetFormField(context.Context, int64, int64) (domain.FormField, error)
	ListSections(context.Context, int64, bool) ([]domain.Section, error)
	CreateSection(context.Context, domain.SectionInput) (int64, error)
	UpdateSection(context.Context, domain.SectionInput) error
	ReorderSections(context.Context, int64, []int64) error
	DeleteSection(context.Context, int64, int64) error
	GetSection(context.Context, int64, int64) (domain.Section, error)
	ListCareers(context.Context, int64, bool) ([]domain.RecruitCareer, error)
	GetCareerBySlug(context.Context, int64, string) (domain.RecruitCareer, error)
	GetCareerByID(context.Context, int64, int64) (domain.RecruitCareer, error)
	CreateCareer(context.Context, domain.RecruitCareerInput) (int64, error)
	UpdateCareer(context.Context, domain.RecruitCareerInput) error
	DeleteCareer(context.Context, int64, int64) error
	GetCustomizationByOwner(context.Context, int64) (domain.Customization, error)
	GetCustomizationBySlug(context.Context, string) (domain.Customization, error)
	UpsertCustomization(context.Context, domain.CustomizationInput) error
	CreateSubmission(context.Context, domain.SubmissionInput) (int64, error)
	ListSubmissions(context.Context, int64, string) ([]domain.Submission, error)
	ListSubmissionsPage(context.Context, int64, string, int32, int32) ([]domain.Submission, error)
	GetSubmissionResumeURL(context.Context, int64, int64) (string, error)
	UpdateSubmissionStatus(context.Context, int64, int64, string) error
}
