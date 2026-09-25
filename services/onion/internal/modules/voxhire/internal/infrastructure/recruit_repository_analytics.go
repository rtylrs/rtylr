package infrastructure

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) loadAnalyticsTotals(ctx context.Context, ownerID int64, totals *domain.AnalyticsTotals) error {
	jobTotals, err := r.queries.AnalyticsCareerJobTotals(ctx, ownerID)
	if err != nil {
		return err
	}
	totals.OpenJobs = ifaceToInt64(jobTotals.OpenJobs)
	totals.ClosedJobs = ifaceToInt64(jobTotals.ClosedJobs)

	appTotals, err := r.queries.AnalyticsApplicationTotals(ctx, ownerID)
	if err != nil {
		return err
	}
	totals.ApplicationsTotal = appTotals.ApplicationsTotal
	totals.Applications7d = ifaceToInt64(appTotals.Applications7d)
	totals.Applications30d = ifaceToInt64(appTotals.Applications30d)
	totals.NewApplications = ifaceToInt64(appTotals.NewApplications)
	cvCount := ifaceToInt64(appTotals.CvCount)
	if totals.ApplicationsTotal > 0 {
		totals.CVRate = float64(cvCount) / float64(totals.ApplicationsTotal)
	}
	if avg, ok := appTotals.AvgSalaryExpectation.(float64); ok {
		totals.AvgSalaryExpectation = avg
	} else if avgBytes, ok := appTotals.AvgSalaryExpectation.([]byte); ok {
		if f, parseErr := strconv.ParseFloat(string(avgBytes), 64); parseErr == nil {
			totals.AvgSalaryExpectation = f
		}
	}

	talentTotals, err := r.queries.AnalyticsTalentPoolTotals(ctx)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.TalentPoolTotal = talentTotals.TalentPoolTotal
		totals.TalentPool7d = ifaceToInt64(talentTotals.TalentPool7d)
		totals.TalentPool30d = ifaceToInt64(talentTotals.TalentPool30d)
	}

	cvTotals, err := r.queries.AnalyticsCVProcessingTotals(ctx, ownerID)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.CVProcessedTotal = ifaceToInt64(cvTotals.CvProcessedTotal)
		totals.CVErrorTotal = ifaceToInt64(cvTotals.CvErrorTotal)
		totals.CVPendingTotal = ifaceToInt64(cvTotals.CvPendingTotal)
	}

	leadTotals, err := r.queries.AnalyticsLeadTotals(ctx)
	if err != nil {
		return err
	}
	totals.LeadsTotal = leadTotals.LeadsTotal
	totals.Leads7d = ifaceToInt64(leadTotals.Leads7d)
	totals.Leads30d = ifaceToInt64(leadTotals.Leads30d)

	admins, err := r.CountAdmins(ctx)
	if err != nil {
		return err
	}
	totals.AdminsTotal = admins

	var taskTotals sqlc.AnalyticsOpenTaskTotalsRow
	err = r.db.QueryRowContext(ctx, analyticsOpenTaskTotalsScopedSQL, ownerID).Scan(
		&taskTotals.OpenTasks, &taskTotals.OverdueTasks,
	)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.OpenTasks = taskTotals.OpenTasks
		totals.OverdueTasks = ifaceToInt64(taskTotals.OverdueTasks)
	}

	var shortlistTotals sqlc.AnalyticsShortlistTotalsRow
	err = r.db.QueryRowContext(ctx, analyticsShortlistTotalsScopedSQL, ownerID).Scan(
		&shortlistTotals.ShortlistsTotal, &shortlistTotals.ShortlistedCandidates,
	)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.ShortlistsTotal = shortlistTotals.ShortlistsTotal
		totals.ShortlistedCandidates = shortlistTotals.ShortlistedCandidates
	}
	return nil
}

func (r *Repository) optionalAnalyticsProfileFunctions(ctx context.Context, ownerID int64) ([]domain.AnalyticsBucket, error) {
	buckets, err := r.analyticsProfileFunctions(ctx, ownerID)
	if err != nil && isOptionalAnalyticsSchemaError(err) {
		return []domain.AnalyticsBucket{}, nil
	}
	return buckets, err
}

func isOptionalAnalyticsSchemaError(err error) bool {
	var myErr *gosqlmysql.MySQLError
	if !errors.As(err, &myErr) {
		return false
	}
	switch myErr.Number {
	case 1054, // unknown column, usually a just-added optional analytics column
		1146: // table does not exist, usually a just-added optional analytics table
		return true
	default:
		return false
	}
}

func (r *Repository) analyticsProfileFunctions(ctx context.Context, ownerID int64) ([]domain.AnalyticsBucket, error) {
	functions := []string{"Operations", "Sales", "Customer Support", "Finance", "HR", "Marketing", "Product", "Design", "Engineering", "Data", "Healthcare", "Retail", "Logistics"}
	selects := make([]string, 0, len(functions))
	args := make([]any, 0, len(functions)+1)
	for _, fn := range functions {
		selects = append(selects, "COALESCE(SUM(profile_json LIKE ?), 0)")
		args = append(args, `%"`+fn+`"%`)
	}
	// cross-schema read from careers — profile function counts across career + talent applications
	args = append(args, ownerID)
	query := `
		SELECT ` + strings.Join(selects, ", ") + `
		  FROM (
		        SELECT COALESCE(a.cv_profile_json, '') AS profile_json
		          FROM voxire_career_applications a
		          JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		        UNION ALL
		        SELECT COALESCE(cv_profile_json, '') AS profile_json
		          FROM voxhire_talent_applications
		       ) profiles`
	counts := make([]int64, len(functions))
	countArgs := make([]any, len(counts))
	for i := range counts {
		countArgs[i] = &counts[i]
	}
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(countArgs...); err != nil {
		return nil, err
	}

	buckets := make([]domain.AnalyticsBucket, 0, len(functions))
	for i, fn := range functions {
		if counts[i] > 0 {
			buckets = append(buckets, domain.AnalyticsBucket{Key: strings.ToLower(strings.ReplaceAll(fn, " ", "_")), Count: counts[i]})
		}
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		if buckets[i].Count == buckets[j].Count {
			return buckets[i].Key < buckets[j].Key
		}
		return buckets[i].Count > buckets[j].Count
	})
	return buckets, nil
}

func (r *Repository) analyticsRecentDays(ctx context.Context, ownerID int64) ([]domain.AnalyticsDay, error) {
	appRows, err := r.queries.AnalyticsApplicationDailyCounts(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	apps := make(map[string]int64, len(appRows))
	for _, row := range appRows {
		apps[row.Day] = ifaceToInt64(row.Count)
	}

	leadRows, err := r.queries.AnalyticsLeadDailyCounts(ctx)
	if err != nil {
		return nil, err
	}
	leads := make(map[string]int64, len(leadRows))
	for _, row := range leadRows {
		leads[row.Day] = ifaceToInt64(row.Count)
	}

	days := make([]domain.AnalyticsDay, 0, 14)
	start := time.Now().UTC().AddDate(0, 0, -13)
	for i := 0; i < 14; i++ {
		key := start.AddDate(0, 0, i).Format("2006-01-02")
		days = append(days, domain.AnalyticsDay{
			Date:         key,
			Applications: apps[key],
			Leads:        leads[key],
		})
	}
	return days, nil
}

// CreateCareer writes a new job posting to the shared careers table.
func (r *Repository) CreateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	slug, err := r.uniqueCareerSlug(ctx, ownerID, in.Title, 0)
	if err != nil {
		return nil, err
	}

	// cross-schema write to careers
	result, err := r.queries.CreateCareer(ctx, sqlc.CreateCareerParams{
		OwnerID: ownerID, Slug: slug, Title: in.Title,
		Department: sqlNullString(in.Department), Location: sqlNullString(in.Location),
		EmploymentType: sqlNullString(in.EmploymentType), WorkMode: sqlNullString(in.WorkMode),
		Seniority: sqlNullString(in.Seniority), Details: in.Details,
		SalaryMin: sqlNullDecimal(in.SalaryMin), SalaryMax: sqlNullDecimal(in.SalaryMax),
		SalaryCurrency: sqlNullString(in.SalaryCurrency),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire career", logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, id)
}
