package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// applyGitHubCodeEffect stores the pull requests and checks a delivery
// carries, in the delivery's transaction.
func applyGitHubCodeEffect(ctx context.Context, tx *sql.Tx, e domain.GitHubWebhookEffect) error {
	for _, pr := range e.PullRequests {
		if _, err := tx.ExecContext(ctx, `INSERT INTO flow_github_pull_requests
  (business_id, repository_id, repo_full_name, number, title, state, head_sha, head_ref, html_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE repo_full_name = VALUES(repo_full_name), title = VALUES(title), state = VALUES(state),
  head_sha = IF(VALUES(head_sha) = '', head_sha, VALUES(head_sha)), head_ref = VALUES(head_ref),
  html_url = VALUES(html_url)`,
			e.BusinessID, pr.RepositoryID, pr.RepoFullName, pr.Number, pr.Title, pr.State, pr.HeadSHA, pr.HeadRef, pr.HTMLURL); err != nil {
			return err
		}
		// The link's state was read once when it was pasted; keep it current.
		if _, err := tx.ExecContext(ctx, `UPDATE flow_task_links SET state = ?
 WHERE business_id = ? AND kind = ? AND ref = ?`,
			pr.State, e.BusinessID, domain.LinkKindGitHubPR, pullRequestRef(pr.RepoFullName, pr.Number)); err != nil {
			return err
		}
	}
	for _, c := range e.Checks {
		// Deliveries can arrive out of order: a late "in_progress" never
		// undoes a completed check. conclusion is set first because MySQL
		// applies the assignments left to right.
		if _, err := tx.ExecContext(ctx, `INSERT INTO flow_github_checks
  (business_id, repository_id, head_sha, external_id, name, status, conclusion, details_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name = VALUES(name),
  conclusion = IF(status = 'completed' AND VALUES(status) <> 'completed', conclusion, VALUES(conclusion)),
  status = IF(status = 'completed' AND VALUES(status) <> 'completed', status, VALUES(status)),
  details_url = IF(VALUES(details_url) = '', details_url, VALUES(details_url))`,
			e.BusinessID, c.RepositoryID, c.HeadSHA, c.ExternalID, c.Name, c.Status, c.Conclusion, c.DetailsURL); err != nil {
			return err
		}
	}
	return nil
}

func pullRequestRef(fullName string, number int64) string {
	return fullName + "#" + strconv.FormatInt(number, 10)
}

// TasksLinkedToGitHubChange returns the tasks whose PR links point at a pull
// request the delivery changed, directly or through its head commit's checks.
func (r *Repository) TasksLinkedToGitHubChange(ctx context.Context, e domain.GitHubWebhookEffect) ([]domain.Task, error) {
	refs := map[string]bool{}
	for _, pr := range e.PullRequests {
		refs[pullRequestRef(pr.RepoFullName, pr.Number)] = true
	}
	for _, c := range e.Checks {
		rows, err := r.pool.QueryContext(ctx, `SELECT repo_full_name, number FROM flow_github_pull_requests
 WHERE business_id = ? AND repository_id = ? AND head_sha = ?`, e.BusinessID, c.RepositoryID, c.HeadSHA)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var name string
			var number int64
			if err := rows.Scan(&name, &number); err != nil {
				_ = rows.Close()
				return nil, err
			}
			refs[pullRequestRef(name, number)] = true
		}
		_ = rows.Close()
	}
	if len(refs) == 0 {
		return nil, nil
	}
	args := []any{e.BusinessID, domain.LinkKindGitHubPR}
	for ref := range refs {
		args = append(args, ref)
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT DISTINCT t.id, t.project_id FROM flow_task_links l
  JOIN flow_tasks t ON t.id = l.task_id AND t.business_id = l.business_id
 WHERE l.business_id = ? AND l.kind = ? AND l.ref IN (`+strings.TrimSuffix(strings.Repeat("?,", len(refs)), ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.ProjectID); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type linkedPR struct {
	taskID       int64
	url, ref     string
	known        bool
	repositoryID int64
	repoFullName string
	number       int64
	title, state string
	headSHA      string
}

// readLinkedPullRequests returns the PR links of the tasks, joined to what the
// webhooks stored about each, in link order.
func (r *Repository) readLinkedPullRequests(ctx context.Context, businessID int64, taskIDs []int64) ([]linkedPR, error) {
	placeholders, ids := int64Placeholders(taskIDs)
	args := append([]any{businessID, domain.LinkKindGitHubPR}, ids...)
	rows, err := r.pool.QueryContext(ctx, `SELECT l.task_id, l.url, COALESCE(l.ref, ''), pr.id IS NOT NULL,
       COALESCE(pr.repository_id, 0), COALESCE(pr.repo_full_name, ''), COALESCE(pr.number, 0),
       COALESCE(pr.title, l.title, ''), COALESCE(pr.state, l.state, ''), COALESCE(pr.head_sha, '')
  FROM flow_task_links l
  LEFT JOIN flow_github_pull_requests pr
    ON pr.business_id = l.business_id AND CONCAT(pr.repo_full_name, '#', pr.number) = l.ref
 WHERE l.business_id = ? AND l.kind = ? AND l.task_id IN (`+placeholders+`)
 ORDER BY l.task_id, l.id`, args...)
	if err != nil {
		logger.Ctx(ctx).Error("readLinkedPullRequests failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []linkedPR
	for rows.Next() {
		var p linkedPR
		if err := rows.Scan(&p.taskID, &p.url, &p.ref, &p.known, &p.repositoryID, &p.repoFullName, &p.number,
			&p.title, &p.state, &p.headSHA); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type shaKey struct {
	repositoryID int64
	sha          string
}

// readChecks returns the latest check per name for each head commit. A
// re-run is a new check run with the same name, and only the newest counts.
func (r *Repository) readChecks(ctx context.Context, businessID int64, keys []shaKey) (map[shaKey][]domain.GitHubCheck, error) {
	out := map[shaKey][]domain.GitHubCheck{}
	if len(keys) == 0 {
		return out, nil
	}
	conds := make([]string, 0, len(keys))
	args := []any{businessID}
	for _, k := range keys {
		conds = append(conds, "(repository_id = ? AND head_sha = ?)")
		args = append(args, k.repositoryID, k.sha)
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT repository_id, head_sha, name, status, conclusion, details_url, updated_at
  FROM flow_github_checks
 WHERE business_id = ? AND (`+strings.Join(conds, " OR ")+`)
 ORDER BY updated_at DESC, id DESC`, args...)
	if err != nil {
		logger.Ctx(ctx).Error("readChecks failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	seen := map[shaKey]map[string]bool{}
	for rows.Next() {
		var k shaKey
		var c domain.GitHubCheck
		var updated time.Time
		if err := rows.Scan(&k.repositoryID, &k.sha, &c.Name, &c.Status, &c.Conclusion, &c.DetailsURL, &updated); err != nil {
			return nil, err
		}
		c.UpdatedAt = updated
		if seen[k] == nil {
			seen[k] = map[string]bool{}
		}
		if seen[k][c.Name] {
			continue
		}
		seen[k][c.Name] = true
		out[k] = append(out[k], c)
	}
	return out, rows.Err()
}

func shaKeysOf(prs []linkedPR) []shaKey {
	var keys []shaKey
	seen := map[shaKey]bool{}
	for _, p := range prs {
		k := shaKey{p.repositoryID, p.headSHA}
		if p.known && p.headSHA != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

// ListTaskPullRequests returns the task's linked PRs with their checks.
func (r *Repository) ListTaskPullRequests(ctx context.Context, businessID, taskID int64) ([]domain.TaskPullRequest, error) {
	prs, err := r.readLinkedPullRequests(ctx, businessID, []int64{taskID})
	if err != nil {
		return nil, err
	}
	checks, err := r.readChecks(ctx, businessID, shaKeysOf(prs))
	if err != nil {
		return nil, err
	}
	out := make([]domain.TaskPullRequest, 0, len(prs))
	for _, p := range prs {
		pr := domain.TaskPullRequest{
			URL: p.url, Ref: p.ref, RepoFullName: p.repoFullName, Number: p.number,
			Title: p.title, State: p.state, Known: p.known, ChecksState: domain.ChecksNone,
		}
		if p.known {
			pr.Checks = checks[shaKey{p.repositoryID, p.headSHA}]
			pr.ChecksState = domain.ChecksState(pr.Checks)
		}
		out = append(out, pr)
	}
	return out, nil
}

// TaskChecksStates returns, per task, the worst check state of its open and
// draft pull requests. Tasks with none are left out.
func (r *Repository) TaskChecksStates(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]string, error) {
	out := map[int64]string{}
	if len(taskIDs) == 0 {
		return out, nil
	}
	prs, err := r.readLinkedPullRequests(ctx, businessID, taskIDs)
	if err != nil {
		return nil, err
	}
	var open []linkedPR
	for _, p := range prs {
		if p.known && (p.state == "open" || p.state == "draft") {
			open = append(open, p)
		}
	}
	checks, err := r.readChecks(ctx, businessID, shaKeysOf(open))
	if err != nil {
		return nil, err
	}
	for _, p := range open {
		state := domain.ChecksState(checks[shaKey{p.repositoryID, p.headSHA}])
		if state == domain.ChecksNone {
			continue
		}
		out[p.taskID] = domain.WorseChecksState(out[p.taskID], state)
	}
	return out, nil
}
