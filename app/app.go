package app

import (
	"context"
	"time"
)

type ContextKey string

const (
	EventUUIDKey             ContextKey    = "event-uuid"
	RepoClonePathKey         ContextKey    = "repo-path"
	BrokerResponseFinished   string        = "finished"
	BrokerResponseTriggered  string        = "triggered"
	BrokerResponseInProgress string        = "in progress"
	BrokerResultSuccess      string        = "success"
	BrokerResultFailure      string        = "failure"
	WorkflowCheckInterval    time.Duration = 10 * time.Second
)

func (ck ContextKey) String() string {
	return string(ck)
}

type GitHubActionsSettings struct {
	GitHubUsername string `yaml:"github_username"`
	GitHubRepo     string `yaml:"github_repo"`
}

type WorkflowResult struct {
	WorkflowID   string
	WorkflowName string
	Status       string
	Result       string
}

// GitHubActions should be implemented to retrieve the GitHub Actions' outcome
type GitHubActions interface {
	GetRepoCommitWorkflowSetup(ctx context.Context, projectID, commitHash string) (*GitHubActionsSettings, error)
	GetRepoCommitWorkflowsResults(ctx context.Context, githubUsername, githubRepo,
		githubCommit string) ([]WorkflowResult, error)
}
