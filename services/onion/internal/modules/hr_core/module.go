package hr_core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/api"
	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/routes"
	workschedulev1 "github.com/voxire/voxie/gen/go/hr/work_schedule/v1"
)

type ApprovalChainInfo struct {
	ID          int64
	RequestType string
	Name        string
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}

type Module struct {
	service *application.Service
}

// Config is the public HR-core configuration shape used by modules that need
// to compose the shared HR configuration endpoint.
type Config struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	CompanyName           string
	FiscalYearStartMonth  int32
	DefaultCurrency       string
	DefaultTimezone       string
	DefaultWorkScheduleID int64
	OvertimeEnabled       bool
	OvertimeMultiplier    float64
	DirectoryEnabled      bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ConfigInput struct {
	CompanyName           string
	FiscalYearStartMonth  int32
	DefaultCurrency       string
	DefaultTimezone       string
	DefaultWorkScheduleID int64
	OvertimeEnabled       bool
	OvertimeMultiplier    float64
	DirectoryEnabled      bool
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) GetConfig(ctx context.Context, ownerID int64) (*Config, error) {
	cfg, err := m.service.GetHRConfig(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return configFromDomain(cfg), nil
}

func (m *Module) UpsertConfig(ctx context.Context, ownerID int64, input ConfigInput) (*Config, error) {
	cfg, err := m.service.UpsertHRConfig(ctx, ownerID, domain.HRCompanyConfigRecord{
		CompanyName:           input.CompanyName,
		FiscalYearStartMonth:  input.FiscalYearStartMonth,
		DefaultCurrency:       input.DefaultCurrency,
		DefaultTimezone:       input.DefaultTimezone,
		DefaultWorkScheduleID: input.DefaultWorkScheduleID,
		OvertimeEnabled:       input.OvertimeEnabled,
		OvertimeMultiplier:    input.OvertimeMultiplier,
		DirectoryEnabled:      input.DirectoryEnabled,
	})
	if err != nil {
		return nil, err
	}
	return configFromDomain(cfg), nil
}

func IsConfigNotFound(err error) bool {
	return errors.Is(err, domain.ErrSetupNotComplete)
}

func configFromDomain(cfg *domain.HRCompanyConfigRecord) *Config {
	return &Config{
		ID:                    cfg.ID,
		OwnerID:               cfg.OwnerID,
		BusinessID:            cfg.BusinessID,
		CompanyName:           cfg.CompanyName,
		FiscalYearStartMonth:  cfg.FiscalYearStartMonth,
		DefaultCurrency:       cfg.DefaultCurrency,
		DefaultTimezone:       cfg.DefaultTimezone,
		DefaultWorkScheduleID: cfg.DefaultWorkScheduleID,
		OvertimeEnabled:       cfg.OvertimeEnabled,
		OvertimeMultiplier:    cfg.OvertimeMultiplier,
		DirectoryEnabled:      cfg.DirectoryEnabled,
		CreatedAt:             cfg.CreatedAt,
		UpdatedAt:             cfg.UpdatedAt,
	}
}

// RegisterAPI registers hr_core's oneKit service servers. AutoRegisterAPIs
// finds this by name on the modules struct, so no wiring in main.go is needed.
//
// hr_core had no oneKit surface before hr.work_schedule.v1: everything else in
// this module is still hand-registered chi routes.
func (m *Module) RegisterAPI(mux *nethttp.ServeMux) {
	if err := workschedulev1.RegisterWorkScheduleServiceServer(
		api.NewWorkScheduleAPI(m.service),
		workschedulev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hr_core work schedule service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hr/employees/invite/verify",
		"/hr/employees/invite/accept",
	}
}

func (m *Module) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return m.service.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

// GetEmployeeBusinessID verifies that an employee belongs to the requested
// business before another module reads employee-owned data.
func (m *Module) GetEmployeeBusinessID(ctx context.Context, ownerID, employeeID int64) (int64, error) {
	employee, err := m.service.GetEmployee(ctx, ownerID, employeeID)
	if err != nil || employee == nil {
		return 0, err
	}
	return employee.BusinessID, nil
}

func (m *Module) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return m.service.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (m *Module) HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error) {
	return m.service.HasRole(ctx, ownerID, employeeID, roleID)
}

// IsDirectoryEnabled reports whether the business has turned on the
// company-wide staff directory. A business with no HR config yet (setup not
// complete) has not opted in, so that case reports false rather than erroring
// - the directory endpoint should deny quietly, not fail the request.
func (m *Module) IsDirectoryEnabled(ctx context.Context, businessID int64) (bool, error) {
	cfg, err := m.service.GetHRConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrSetupNotComplete) {
			return false, nil
		}
		return false, err
	}
	return cfg.DirectoryEnabled, nil
}

// EmployeeSummary is the minimal shape another module needs to resolve who
// an employee is - name and which hub user they're linked to - without
// reaching into hr_core's internal domain package.
type EmployeeSummary struct {
	ID        int64
	UserID    int64
	FirstName string
	LastName  string
}

// ListActiveEmployees returns every active or probationary employee for a
// business, paginating internally against ListEmployees' own 100-row page
// cap so a caller (Flow's business-wide "who's in today") gets the whole
// roster in one call rather than reimplementing paging itself. maxPages is a
// runaway-loop safety bound, not a real product limit.
func (m *Module) ListActiveEmployees(ctx context.Context, ownerID int64) ([]EmployeeSummary, error) {
	const pageSize = 100
	const maxPages = 50

	var all []EmployeeSummary
	for page := int32(0); page < maxPages; page++ {
		employees, total, err := m.service.ListEmployees(ctx, ownerID, pageSize, page*pageSize)
		if err != nil {
			return nil, err
		}
		for _, e := range employees {
			if e.Status != "active" && e.Status != "probation" {
				continue
			}
			all = append(all, EmployeeSummary{ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName})
		}
		if int64((page+1)*pageSize) >= total || len(employees) == 0 {
			break
		}
	}
	return all, nil
}

func (m *Module) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error) {
	chain, err := m.service.GetApprovalChain(ctx, ownerID, requestType)
	if err != nil {
		return nil, err
	}
	if chain == nil {
		return nil, nil
	}
	levels := make([]ApprovalLevelInfo, 0, len(chain.Levels))
	for _, lv := range chain.Levels {
		levels = append(levels, ApprovalLevelInfo{
			LevelOrder:         lv.LevelOrder,
			ApproverType:       lv.ApproverType,
			ApproverRoleID:     lv.ApproverRoleID,
			ApproverEmployeeID: lv.ApproverEmployeeID,
		})
	}
	return &ApprovalChainInfo{
		ID:          chain.ID,
		RequestType: chain.RequestType,
		Name:        chain.Name,
		TotalLevels: int32(len(chain.Levels)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Levels:      levels,
	}, nil
}

// SetTerminationRevoker connects employment termination to the entitlement
// owner without importing it: the port lives in hr_core's application layer
// and team's service satisfies it structurally (#834).
func (m *Module) SetTerminationRevoker(r application.TerminationRevoker) {
	m.service.SetTerminationRevoker(r)
}
