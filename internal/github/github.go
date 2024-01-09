package github

import (
	"context"
	"github.com/google/go-github/v57/github"
	"log/slog"
	"radicle-github-actions-adapter/app/githubops"
	"strconv"
)

type GitHub struct {
	logger  *slog.Logger
	pat     string
	repos   RepositoriesService
	actions ActionsService
}

type RepositoriesService interface {
	GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit,
		*github.Response, error)
}

type ActionsService interface {
	ListRepositoryWorkflowRuns(ctx context.Context, owner, repo string,
		opts *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error)
}

func NewGitHub(pat string, logger *slog.Logger) *GitHub {
	ghClient := github.NewClient(nil).WithAuthToken(pat)
	return &GitHub{
		logger:  logger,
		repos:   ghClient.Repositories,
		actions: ghClient.Actions,
	}
}

// CheckRepoCommit checks if repo and commit are present from GitHub
func (gh *GitHub) CheckRepoCommit(ctx context.Context, user, repo, commit string) error {
	_, _, err := gh.repos.GetCommit(ctx, user, repo, commit, nil)
	if err != nil {
		gh.logger.Error("failed to get repo commit", "error", err.Error())
		return err
	}
	return nil
}

// GetRepoCommitWorkflows returns all the available workflows of the specified repo and commit.
// If no workflows exist it does not return any error.
func (gh *GitHub) GetRepoCommitWorkflows(ctx context.Context, user, repo, commit string) ([]githubops.WorkflowResult, error) {
	ListOptions := github.ListOptions{
		Page:    0,
		PerPage: 30, //default 30, range [0-100]
	}
	var result []githubops.WorkflowResult
	for {
		runs, resp, err := gh.actions.ListRepositoryWorkflowRuns(ctx, user, repo,
			&github.ListWorkflowRunsOptions{
				HeadSHA:     commit,
				ListOptions: ListOptions,
			})
		if err != nil {
			gh.logger.Error("failed to get repo commit", "error", err.Error())
			return nil, err
		}

		for _, run := range runs.WorkflowRuns {
			result = append(result, githubops.WorkflowResult{
				WorkflowID:   strconv.Itoa(int(run.GetID())),
				WorkflowName: run.GetName(),
				Status:       run.GetStatus(),
				Result:       run.GetConclusion(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		ListOptions.Page = resp.NextPage
	}
	return result, nil
}
