package githubops

import "context"

const (
	WorkflowResultSuccess string = "success"
	WorkflowResultFailure string = "failure"

	WorkflowStatusCompleted  string = "completed"
	WorkflowStatusInProgress string = "in_progress"
)

type WorkflowResult struct {
	WorkflowID   string
	WorkflowName string
	Status       string
	Result       string
}

type GitHubOps interface {
	CheckRepoCommit(ctx context.Context, user, repo, commit string) error
	GetRepoCommitWorkflows(ctx context.Context, user, repo, commit string) ([]WorkflowResult, error)
}
