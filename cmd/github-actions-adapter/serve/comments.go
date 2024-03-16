package serve

import (
	"context"
	"errors"
	"fmt"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
	"radicle-github-actions-adapter/app/githubops"
)

// commentResultOnPatch adds a patch-revision comment with the results of the GitHub workflows.
func (gas *GitHubActionsServer) commentOnPatch(ctx context.Context,
	brokerRequestMessage *broker.RequestMessage, commentMessage string) error {
	if len(brokerRequestMessage.PatchEvent.Patch.Revisions) == 0 {
		gas.App.Logger.Warn("could not comment on patch", "error", "no revision found in patch")
		return errors.New("no revision found in patch")
	}
	revision := brokerRequestMessage.PatchEvent.Patch.Revisions[len(brokerRequestMessage.PatchEvent.Patch.Revisions)-1]
	err := gas.Radicle.Comment(ctx, brokerRequestMessage.Repo, brokerRequestMessage.PatchEvent.Patch.ID, revision.ID,
		commentMessage)
	if err != nil {
		gas.App.Logger.Warn("could not comment on patch", "content", commentMessage, "patch_id",
			brokerRequestMessage.PatchEvent.Patch.ID, "revision_id", revision.ID, "error", err.Error())
		return err
	}
	gas.App.Logger.Debug("successfully added patch comment", "content", commentMessage, "patch_id",
		brokerRequestMessage.PatchEvent.Patch.ID, "revision_id", revision.ID)
	return nil

}

// preparePatchCommentStartMessage prepares a message for adding as patch comment with information about starting to
// check for GitHub workflows.
func (gas *GitHubActionsServer) preparePatchCommentStartMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	commentMessage := "Checking for GitHub Actions Workflows."
	return commentMessage
}

// preparePatchCommentInfoMessage prepares a message for adding as patch comment with information about the GitHub
// workflows.
func (gas *GitHubActionsServer) preparePatchCommentInfoMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	githubWorkflowURL := "https://github.com/%s/%s/actions/runs/%s"
	commentMessage := "GitHub Actions Workflows ⏳"

	commentMessage += "\n\nWorkflows:"
	for _, result := range resultResponse.ResultDetails {
		url := fmt.Sprintf(githubWorkflowURL, gitHubActionsSettings.GitHubUsername, gitHubActionsSettings.GitHubRepo, result.WorkflowID)
		commentMessage += "\n\n - "
		commentMessage += fmt.Sprintf(`[%s (%s) ⏳](%s "started")`, result.WorkflowName, result.WorkflowID, url)
	}
	return commentMessage
}

// preparePatchCommentResultMessage prepares a message for adding as patch comment with the workflow results.
func (gas *GitHubActionsServer) preparePatchCommentResultMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	githubWorkflowURL := "https://github.com/%s/%s/actions/runs/%s"
	commentMessage := "GitHub Actions Result: " + resultResponse.Result
	if resultResponse.Result == app.BrokerResultSuccess {
		commentMessage += " ✅"
	} else {
		commentMessage += " ❌"
	}

	commentMessage += "\n\nDetails:"
	for _, result := range resultResponse.ResultDetails {
		url := fmt.Sprintf(githubWorkflowURL, gitHubActionsSettings.GitHubUsername, gitHubActionsSettings.GitHubRepo, result.WorkflowID)
		commentMessage += "\n\n - "
		icon := "⚠️️"
		if result.WorkflowResult == githubops.WorkflowResultSuccess {
			icon = "✅"
		} else if result.WorkflowResult == githubops.WorkflowResultFailure {
			icon = "❌"
		}
		commentMessage += fmt.Sprintf(`[%s (%s) %s](%s "%s")`, result.WorkflowName, result.WorkflowID, icon, url,
			result.WorkflowResult)
	}
	return commentMessage
}
