package api

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowgithubv2 "github.com/voxire/voxie/gen/go/flow/github/v2"
)

func (a *GitHubAPIV2) ListTaskPullRequests(ctx context.Context, req *flowgithubv2.ListTaskPullRequestsRequest) (*flowgithubv2.ListTaskPullRequestsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	prs, err := a.Service.ListTaskPullRequests(ctx, caller.UserID, businessID, req.GetTaskId())
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "ListTaskPullRequests failed", err)
	}
	out := make([]*flowgithubv2.TaskPullRequest, 0, len(prs))
	for _, pr := range prs {
		checks := make([]*flowgithubv2.GitHubCheck, 0, len(pr.Checks))
		for _, c := range pr.Checks {
			checks = append(checks, &flowgithubv2.GitHubCheck{
				Name: c.Name, Status: c.Status, Conclusion: c.Conclusion, DetailsUrl: c.DetailsURL,
				UpdatedAt: c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			})
		}
		out = append(out, &flowgithubv2.TaskPullRequest{
			Url: pr.URL, Ref: pr.Ref, Title: pr.Title, State: pr.State, Known: pr.Known,
			ChecksState: pr.ChecksState, Checks: checks,
		})
	}
	return &flowgithubv2.ListTaskPullRequestsResponse{PullRequests: out}, nil
}
