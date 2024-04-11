package radiclegithubactions

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/githubops"
	"radicle-github-actions-adapter/app/gitops"
	"strings"
)

const (
	radicleGitHubActionsSettingsPath string = "/.radicle/github_actions.yaml"
	gitHubActionsWorkflowsPath       string = "/.github/workflows"
)

type RadicleGitHubActions struct {
	logger      *slog.Logger
	radicleHome string
	git         gitops.GitOps
	github      githubops.GitHubOps
}

func NewRadicleGitHubActions(radicleHome string, gitOps gitops.GitOps, githubOps githubops.GitHubOps,
	logger *slog.Logger) *RadicleGitHubActions {
	return &RadicleGitHubActions{
		logger:      logger,
		radicleHome: radicleHome,
		git:         gitOps,
		github:      githubOps,
	}
}

// GetRepoCommitWorkflowSetup returns the GitHub Actions setup if any.
// The setup is located at gitHubActionsWorkflowsPath path.
// Checks also if there are registered workflows under radicleGitHubActionsSettingsPath
func (rga *RadicleGitHubActions) GetRepoCommitWorkflowSetup(ctx context.Context, projectID,
	commitHash string) (*app.GitHubActionsSettings, error) {
	repoPath := ctx.Value(app.RepoClonePathKey).(string)
	projectID = strings.TrimPrefix(projectID, "rad:")
	cloneURL := fmt.Sprintf("file://%s/storage/%s", rga.radicleHome, projectID)
	defer os.RemoveAll(repoPath)

	rga.logger.Info("cloning project", "ID", projectID, "url", cloneURL, "to", repoPath)
	err := rga.git.CloneRepoCommit(cloneURL, commitHash, repoPath)
	if err != nil {
		rga.logger.Error("failed to clone repo from URL", "url", cloneURL, "error", err.Error())
		return nil, err
	}

	githubActionsSetup, err := rga.getRadicleGitHubActionsSetup(repoPath + radicleGitHubActionsSettingsPath)
	if err != nil {
		rga.logger.Warn("no GitHub Actions setup found", "reason", err.Error())
		return nil, nil
	}
	if githubActionsSetup == nil || len(githubActionsSetup.GitHubUsername) == 0 || len(githubActionsSetup.GitHubRepo) == 0 {
		rga.logger.Warn("empty GitHub Actions setup found")
		return nil, nil
	}

	githubActionsYamlFilePaths, err := rga.listYAMLFiles(repoPath + gitHubActionsWorkflowsPath)
	if err != nil || len(githubActionsYamlFilePaths) == 0 {
		rga.logger.Warn("no GitHub Actions workflows files found")
		return nil, nil
	}
	rga.logger.Debug(fmt.Sprintf("found GitHub actions workflows yaml files: %+v", githubActionsYamlFilePaths))
	return githubActionsSetup, nil
}

// GetRepoCommitWorkflowsResults retrieves the repo's workflows results from GitHub.
func (rga *RadicleGitHubActions) GetRepoCommitWorkflowsResults(ctx context.Context, githubUsername, githubRepo,
	githubCommit string) ([]app.WorkflowResult, error) {
	err := rga.github.CheckRepoCommit(ctx, githubUsername, githubRepo, githubCommit)
	if err != nil {
		rga.logger.Error("no GitHub repo commit found", "error", err.Error())
		return nil, err
	}
	githubWorkflows, err := rga.github.GetRepoCommitWorkflows(ctx, githubUsername, githubRepo, githubCommit)
	if err != nil {
		rga.logger.Error("could not check for GitHub workflows", "error", err.Error())
		return nil, err
	}
	var workflows []app.WorkflowResult
	for _, githubWorkflow := range githubWorkflows {
		var workflowArtifacts []app.WorkflowArtifact
		for _, artifact := range githubWorkflow.Artifacts {
			workflowArtifacts = append(workflowArtifacts, app.WorkflowArtifact{
				Id:     artifact.Id,
				Name:   artifact.Name,
				Url:    artifact.Url,
				ApiUrl: artifact.ApiUrl,
			})
		}
		workflows = append(workflows, app.WorkflowResult{
			WorkflowID:   githubWorkflow.WorkflowID,
			WorkflowName: githubWorkflow.WorkflowName,
			Status:       githubWorkflow.Status,
			Result:       githubWorkflow.Result,
			Artifacts:    workflowArtifacts,
		})
	}
	rga.logger.Debug(fmt.Sprintf("found GitHub actions workflows: %+v", workflows))
	return workflows, nil
}
