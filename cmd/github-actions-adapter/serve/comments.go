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
	brokerRequestMessage *broker.RequestMessage, commentMessage string, append bool) error {
	if len(brokerRequestMessage.PatchEvent.Patch.Revisions) == 0 {
		gas.App.Logger.Warn("could not comment on patch", "error", "no revision found in patch")
		return errors.New("no revision found in patch")
	}
	revision := brokerRequestMessage.PatchEvent.Patch.Revisions[len(brokerRequestMessage.PatchEvent.Patch.Revisions)-1]
	err := gas.Radicle.Comment(ctx, brokerRequestMessage.Repo, brokerRequestMessage.PatchEvent.Patch.ID, revision.ID,
		commentMessage, append)
	if err != nil {
		gas.App.Logger.Warn("could not comment on patch", "content", commentMessage, "patch_id",
			brokerRequestMessage.PatchEvent.Patch.ID, "revision_id", revision.ID, "error", err.Error())
		return err
	}
	gas.App.Logger.Debug("successfully added patch comment", "content", commentMessage, "patch_id",
		brokerRequestMessage.PatchEvent.Patch.ID, "revision_id", revision.ID)
	return nil

}

// preparePatchCommentResultMessage prepares a message for adding as patch comment with the workflow results.
func (gas *GitHubActionsServer) preparePatchCommentResultMessage(resultResponse broker.ResponseMessage,
	gitHubActionsSettings app.GitHubActionsSettings) string {
	githubWorkflowURL := "https://github.com/%s/%s/actions/runs/%s"
	actionsStatus := "Result"
	if resultResponse.Response == app.BrokerResponseInProgress {
		actionsStatus = "Status"
	}
	commentMessage := "### GitHub Actions " + actionsStatus + ": "
	if resultResponse.Response == app.BrokerResponseInProgress {
		commentMessage += app.BrokerResponseInProgress
		commentMessage += " ⏳"
	} else {
		commentMessage += resultResponse.Result
		if resultResponse.Result == app.BrokerResultSuccess {
			commentMessage += " ✅"
		} else {
			commentMessage += " ❌"
		}
	}
	commentMessage += "\n Workflows:"
	for _, result := range resultResponse.ResultDetails {
		url := fmt.Sprintf(githubWorkflowURL, gitHubActionsSettings.GitHubUsername, gitHubActionsSettings.GitHubRepo,
			result.WorkflowID)
		commentMessage += "\n - "
		icon := "⚠️️"
		if result.WorkflowResult == githubops.WorkflowStatusInProgress {
			icon = "⏳"
		} else if result.WorkflowResult == githubops.WorkflowResultSuccess {
			icon = "✅"
		} else if result.WorkflowResult == githubops.WorkflowResultFailure {
			icon = "❌"
		}
		commentMessage += fmt.Sprintf(`[%s (%s) %s](%s "%s")`, result.WorkflowName, result.WorkflowID, icon, url,
			result.WorkflowResult)
		if len(result.WorkflowArtifacts) > 0 {
			commentMessage += "\n\t Artifacts:"
			for _, artifact := range result.WorkflowArtifacts {
				commentMessage += fmt.Sprintf("\n\t\t - [%s (%s)](%s)", artifact.Name, artifact.Id, artifact.Url)
			}
		}
	}
	return commentMessage
}
