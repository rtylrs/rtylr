package application

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func (s *Service) CreateCycleV2(ctx context.Context, businessID, projectID int64, name, startDate, endDate string) (domain.Cycle, error) {
	sd, ed, dateErr := NormalizeCycleDates(startDate, endDate)
	if dateErr != nil {
		return domain.Cycle{}, dateErr
	}
	status := "upcoming"
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if today.After(ed) {
		status = "completed"
	} else if !today.Before(sd) {
		status = "active"
	}
	id, err := s.repo.CreateCycleV2(ctx, domain.CycleInput{
		BusinessID: businessID, ProjectID: &projectID,
		Name: name, StartDate: sd, EndDate: ed, Status: status,
	})
	if err != nil {
		return domain.Cycle{}, err
	}
	// The normalized dates go back out with the cycle. Without them the create
	// response carries no dates, so a client that renders straight from it has
	// to re-read the row to learn when the cycle runs.
	return domain.Cycle{
		ID: id, ProjectID: &projectID, Name: name, Status: status,
		StartDate: sd, EndDate: ed,
	}, nil
}

func (s *Service) ListMilestonesV2(ctx context.Context, businessID int64, projectID *int64, callerID, limit, offset int64) ([]domain.Milestone, int64, error) {
	return s.repo.ListMilestonesV2(ctx, businessID, projectID, callerID, false, limit, offset)
}

// UpdateMilestoneV2 writes the target_date column the v2 list reads; the v1
// update writes due_date, the column v1 reads.
func (s *Service) UpdateMilestoneV2(ctx context.Context, input domain.UpdateMilestoneInput) error {
	return s.repo.UpdateMilestoneV2(ctx, input)
}

func (s *Service) CreateMilestoneV2(ctx context.Context, businessID, projectID int64, name, description, targetDate string) (domain.Milestone, error) {
	var dueDate time.Time
	if targetDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", targetDate)
		if parseErr != nil {
			return domain.Milestone{}, ErrInvalidMilestoneDate
		}
		dueDate = parsed
	}

	id, err := s.repo.CreateMilestoneV2(ctx, domain.MilestoneInput{
		BusinessID: businessID, ProjectID: projectID,
		Name: name, Description: description, DueDate: dueDate,
	})
	if err != nil {
		return domain.Milestone{}, err
	}
	return domain.Milestone{ID: id, ProjectID: projectID, Name: name, Description: description, Status: "planning", DueDate: dueDate}, nil
}

// callerID drives the private-view filter: a private view is visible only to
// whoever created it.
func (s *Service) ListViewsV2(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViewsV2(ctx, businessID, callerID, limit, offset)
}

func (s *Service) CreateViewV2(ctx context.Context, businessID, createdBy int64, name, icon, filtersJSON, groupBy, sortBy, sortDir, visibility string) (domain.View, error) {
	var filters map[string]interface{}
	_ = json.Unmarshal([]byte(filtersJSON), &filters)
	id, err := s.repo.CreateViewV2(ctx, domain.ViewInput{
		BusinessID: businessID, Name: name, Icon: icon, Filters: filters,
		GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir, CreatedBy: createdBy,
		Visibility: visibility,
	})
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{ID: id, Name: name, Icon: icon, Filters: filters, GroupBy: groupBy, SortBy: sortBy, SortDir: sortDir, Visibility: visibility}, nil
}

// UpdateViewV2 carries the whole view, because repo.UpdateView is a REPLACEMENT:
// its statement sets name, icon, filters, group_by, sort_by, sort_dir and
// visibility unconditionally. The previous signature took only the id, so every
// v2 update wrote empty strings over all of them - renaming a view through v2
// blanked it.
func (s *Service) UpdateViewV2(ctx context.Context, actorID, businessID, viewID int64, name, icon, filtersJSON, groupBy, sortBy, sortDir, visibility string) error {
	var filters map[string]interface{}
	_ = json.Unmarshal([]byte(filtersJSON), &filters)
	return s.repo.UpdateView(ctx, domain.UpdateViewInput{
		ID:         viewID,
		ActorID:    actorID,
		BusinessID: businessID,
		Name:       name,
		Icon:       icon,
		Filters:    filters,
		GroupBy:    groupBy,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Visibility: visibility,
	})
}

// profilePublicProjectCap bounds the profile card list. A profile is a summary,
// not a paginated project browser, and All projects already exists for the
// full list.
const profilePublicProjectCap = 24

// GetFlowProfile returns one business member's profile: identity from the
// HR-owned hub profile, plus the public projects they created.
//
// Cross-business isolation comes from the resolver: it looks the subject up
// inside the caller's business, so a subject who is not a colleague returns
// ErrUserNotFound rather than a profile. There is deliberately no extra role
// gate on top - a colleague's name, email and public projects are already
// visible through the member roster and the project list.
func (s *Service) GetFlowProfile(
	ctx context.Context,
	subjectID, businessID int64,
) (domain.FlowProfile, error) {
	if s.hubProfiles == nil {
		return domain.FlowProfile{}, domain.ErrUnauthorized
	}
	hub, err := s.hubProfiles.GetProfile(ctx, subjectID, businessID)
	if err != nil {
		return domain.FlowProfile{}, err
	}

	projects, err := s.repo.ListPublicProjectsByOwnerV2(ctx, businessID, subjectID, profilePublicProjectCap)
	if err != nil {
		return domain.FlowProfile{}, err
	}

	return domain.FlowProfile{
		UserID:         hub.ID,
		Name:           hub.FullName,
		Email:          hub.Email,
		Initials:       initialsOf(hub.FullName),
		RoleTitle:      "",
		Bio:            "",
		PublicProjects: projects,
	}, nil
}

// initialsOf builds the two-letter monogram the contract asks for.
//
// Falls back to the first two letters of a single-word name, and to the empty
// string for an empty name, so the caller never has to special-case it.
func initialsOf(name string) string {
	fields := strings.Fields(name)
	switch len(fields) {
	case 0:
		return ""
	case 1:
		one := []rune(fields[0])
		if len(one) == 1 {
			return strings.ToUpper(string(one[0]))
		}
		return strings.ToUpper(string(one[0:2]))
	default:
		first := []rune(fields[0])
		last := []rune(fields[len(fields)-1])
		return strings.ToUpper(string(first[0]) + string(last[0]))
	}
}
