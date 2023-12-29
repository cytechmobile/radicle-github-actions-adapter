package app

import (
	"context"
	"time"
)

const (
	EventUUIDKey            string        = "event-uuid"
	RepoClonePath           string        = "repo-path"
	BrokerResponseFinished  string        = "finished"
	BrokerResponseTriggered string        = "triggered"
	BrokerResultSuccess     string        = "success"
	BrokerResultFailure     string        = "failure"
	WorkflowCheckInterval   time.Duration = 10 * time.Second
)

type GitHubActionsSettings struct {
	GitHubUsername string `yaml:"github_username"`
	GitHubRepo     string `yaml:"github_repo"`
}

type WorkflowResult struct {
	WorkflowID string
	Status     string
	Result     string
}

// GitHubActions should be implemented to retrieve the GitHub Actions' outcome
type GitHubActions interface {
	GetRepoCommitWorkflowSetup(ctx context.Context, projectID, commitHash string) (*GitHubActionsSettings, error)
	GetRepoCommitWorkflows(ctx context.Context, githubUsername, githubRepo,
		githubCommit string) ([]WorkflowResult, error)
}
