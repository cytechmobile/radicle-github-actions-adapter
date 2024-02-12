package serve

import (
	"context"
	"errors"
	"fmt"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
	"radicle-github-actions-adapter/app/githubops"
)

// commentResultOnPatch adds a patch-revision comment with the results of the Github workflows.
func (gas *GitHubActionsServer) commentOnPatch(ctx context.Context,
	brokerRequestMessage *broker.RequestMessage, commentMessage string) error {
	if len(brokerRequestMessage.PatchEvent.Patch.Revisions) == 0 {
		gas.App.Logger.Warn("could not comment on patch", "error", "no revision found in patch")
		return errors.New("no revision found in patch")
	}
	revision := brokerRequestMessage.PatchEvent.Patch.Revisions[len(brokerRequestMessage.PatchEvent.Patch.Revisions)-1]
	return gas.Radicle.Comment(ctx, brokerRequestMessage.Repo, brokerRequestMessage.PatchEvent.Patch.ID, revision.ID,
		commentMessage)
}

// preparePatchCommentStartMessage prepares a message for adding as patch comment with information about starting to
// check for Github workflows.
func (gas *GitHubActionsServer) preparePatchCommentStartMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	commentMessage := "Checking for Github Actions Workflows."
	return commentMessage
}

// preparePatchCommentInfoMessage prepares a message for adding as patch comment with information about the Github
// workflows.
func (gas *GitHubActionsServer) preparePatchCommentInfoMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	githubWorkflowURL := "https://github.com/%s/%s/actions/runs/%s"
	commentMessage := "Github Actions Workflows 🟧"

	commentMessage += "\n\nWorkflows:"
	for _, result := range resultResponse.ResultDetails {
		url := fmt.Sprintf(githubWorkflowURL, gitHubActionsSettings.GitHubUsername, gitHubActionsSettings.GitHubRepo, result.WorkflowID)
		commentMessage += "\n\n - "
		commentMessage += `<a href="` + url + `" target="_blank" >` + result.WorkflowName + " (" + result.
			WorkflowID + ")</a> 🟠"
	}
	return commentMessage
}

// preparePatchCommentResultMessage prepares a message for adding as patch comment with the workflow results.
func (gas *GitHubActionsServer) preparePatchCommentResultMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	githubWorkflowURL := "https://github.com/%s/%s/actions/runs/%s"
	commentMessage := "Github Actions Result: " + resultResponse.Result
	if resultResponse.Result == githubops.WorkflowResultSuccess {
		commentMessage += " ✅"
	} else {
		commentMessage += " ❌"
	}

	commentMessage += "\n\nDetails:"
	for _, result := range resultResponse.ResultDetails {
		url := fmt.Sprintf(githubWorkflowURL, gitHubActionsSettings.GitHubUsername, gitHubActionsSettings.GitHubRepo, result.WorkflowID)
		commentMessage += "\n\n - "
		commentMessage += `<a href="` + url + `" target="_blank" >` + result.WorkflowName + " (" + result.
			WorkflowID + ")</a>: " + result.WorkflowResult
		if result.WorkflowResult == githubops.WorkflowResultSuccess {
			commentMessage += " 🟢"
		} else if result.WorkflowResult == githubops.WorkflowResultFailure {
			commentMessage += " 🔴"
		} else {
			commentMessage += " 🟡"
		}
	}
	return commentMessage
}
