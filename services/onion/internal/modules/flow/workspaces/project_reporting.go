package workspaces

import "context"

// The two readers below let another module ask Flow who may see a project's
// reporting without reimplementing Flow's privacy rules or reading flow_*
// tables itself. The project dashboard in insights is the consumer.

// CanSeeProjectReport reports whether userID may see reporting on projectID in
// businessID: anyone Flow lets open the project, plus anyone with Flow's
// business-wide reporting visibility. Personal My Tasks projects are never
// reported on. A missing project is false with no error.
func (m *Module) CanSeeProjectReport(ctx context.Context, businessID, userID, projectID int64) (bool, error) {
	return m.service.CanSeeProjectReport(ctx, projectID, userID, businessID)
}

// ListedProjectIDs returns the ids of the projects Flow's sidebar lists for
// userID in businessID: curated for everyone, bypass roles included.
func (m *Module) ListedProjectIDs(ctx context.Context, businessID, userID int64) ([]int64, error) {
	return m.service.ListedProjectIDs(ctx, userID, businessID)
}
