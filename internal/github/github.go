package github

import (
	"context"
	"fmt"
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
	ListWorkflowRunArtifacts(ctx context.Context, owner, repo string, runID int64,
		opts *github.ListOptions) (*github.ArtifactList, *github.Response, error)
}

func NewGitHub(pat string, logger *slog.Logger) *GitHub {
	ghClient := &github.Client{}
	if len(pat) == 0 {
		ghClient = github.NewClient(nil)
	} else {
		ghClient = github.NewClient(nil).WithAuthToken(pat)
	}
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
	workflowsListOptions := github.ListOptions{
		Page:    0,
		PerPage: 30, //default 30, range [0-100]
	}
	var result []githubops.WorkflowResult
	for {
		runs, workflowsResp, err := gh.actions.ListRepositoryWorkflowRuns(ctx, user, repo,
			&github.ListWorkflowRunsOptions{
				HeadSHA:     commit,
				ListOptions: workflowsListOptions,
			})
		if err != nil {
			gh.logger.Error("failed to get repo commit", "error", err.Error())
			return nil, err
		}
		for _, run := range runs.WorkflowRuns {
			artifactsListOptions := github.ListOptions{
				Page:    0,
				PerPage: 30, //default 30, range [0-100]
			}
			var workflowArtifacts []github.Artifact
			for {
				artifacts, artifactsResp, err := gh.actions.ListWorkflowRunArtifacts(ctx, user, repo, run.GetID(),
					&artifactsListOptions)
				if err != nil {
					gh.logger.Error("could not fetch workflow artifacts", "error", err.Error())
					break
				}
				for _, artifact := range artifacts.Artifacts {
					workflowArtifacts = append(workflowArtifacts, *artifact)
				}
				if artifactsResp.NextPage == 0 {
					break
				}
			}
			var resultArtifacts []githubops.WorkflowArtifact
			for _, workflowArtifact := range workflowArtifacts {
				resultArtifacts = append(resultArtifacts, githubops.WorkflowArtifact{
					Id:   strconv.FormatInt(workflowArtifact.GetID(), 10),
					Name: workflowArtifact.GetName(),
					Url: fmt.Sprintf("https://github.com/%s/%s/actions/runs/%d/artifacts/%d",
						user, repo, run.GetID(), workflowArtifact.GetID()),
					ApiUrl: workflowArtifact.GetURL(),
				})
			}
			result = append(result, githubops.WorkflowResult{
				WorkflowID:   strconv.FormatInt(run.GetID(), 10),
				WorkflowName: run.GetName(),
				Status:       run.GetStatus(),
				Result:       run.GetConclusion(),
				Artifacts:    resultArtifacts,
			})
			artifactsListOptions.Page++
		}
		if workflowsResp.NextPage == 0 {
			break
		}
		workflowsListOptions.Page = workflowsResp.NextPage
	}
	return result, nil
}
