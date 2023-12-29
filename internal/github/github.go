package github

import (
	"context"
	"github.com/google/go-github/v57/github"
	"log/slog"
	"radicle-github-actions-adapter/app/githubops"
	"strconv"
)

type GitHub struct {
	logger *slog.Logger
	pat    string
	client *github.Client
}

func NewGitHub(pat string, logger *slog.Logger) *GitHub {
	return &GitHub{
		logger: logger,
		client: github.NewClient(nil).WithAuthToken(pat),
	}
}

// CheckRepoCommit checks if repo and commit are present from GitHub
func (gh *GitHub) CheckRepoCommit(ctx context.Context, user, repo, commit string) error {
	_, _, err := gh.client.Repositories.GetCommit(ctx, user, repo, commit, nil)
	if err != nil {
		gh.logger.Error("failed to get repo commit", "error", err.Error())
		return err
	}
	return nil
}

// GetRepoCommitWorkflows returns all the available workflows of the specified repo and commit.
// If no workflows exist it does not return any error.
func (gh *GitHub) GetRepoCommitWorkflows(ctx context.Context, user, repo, commit string) ([]githubops.WorkflowResult, error) {
	ListOptions := github.ListOptions{}
	var result []githubops.WorkflowResult
	for {
		runs, resp, err := gh.client.Actions.ListRepositoryWorkflowRuns(ctx, user, repo,
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
				WorkflowID: strconv.Itoa(int(run.GetID())),
				Status:     run.GetStatus(),
				Result:     run.GetConclusion(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		ListOptions.Page = resp.NextPage
	}
	return result, nil
}
