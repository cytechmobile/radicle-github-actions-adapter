package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
	"radicle-github-actions-adapter/app/githubops"
	"radicle-github-actions-adapter/app/radicle"
	"time"
)

type AppConfig struct {
	RadicleHome             string
	GitHubPAT               string
	WorkflowsPollTimoutMins uint64
	RadicleHttpdURL         string
	RadicleSessionToken     string
}

type App struct {
	Config AppConfig
	Logger *slog.Logger
}

// GitHubActionsServer is used as a container for the most important dependencies.
type GitHubActionsServer struct {
	App           *App
	Broker        broker.Broker
	GitHubActions app.GitHubActions
	Radicle       radicle.Patch
}

// NewGitHubActionsServer returns a pointer to a new GitHub Action Server.
func NewGitHubActionsServer(config *App, broker broker.Broker,
	GitHubActions app.GitHubActions, radiclePatrch radicle.Patch) *GitHubActionsServer {
	server := &GitHubActionsServer{
		App:           config,
		Broker:        broker,
		GitHubActions: GitHubActions,
		Radicle:       radiclePatrch,
	}
	return server
}

// Serve is responsible for parsing stdin input and check any GitHub Actions status of radicle projects
// It also manages replies to the broker.
func (gas *GitHubActionsServer) Serve(ctx context.Context) error {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	gas.App.Logger.Info("serving event", app.EventUUIDKey.String(), eventUUID)
	brokerRequestMessage, err := gas.Broker.ParseRequestMessage(ctx)
	if err != nil {
		gas.App.Logger.Error(err.Error())
		return err
	}
	gas.App.Logger.Debug("received message type", "message", fmt.Sprintf("%+v", *brokerRequestMessage))

	jobResponse := broker.ResponseMessage{
		Response: app.BrokerResponseTriggered,
		RunID: &broker.RunID{
			ID: eventUUID,
		},
	}
	gas.App.Logger.Debug("sending message", "message", fmt.Sprintf("%+v", jobResponse))
	err = gas.Broker.ServeResponse(ctx, jobResponse)
	if err != nil {
		gas.App.Logger.Error(err.Error())
		return err
	}

	repoCommitWorkflowSetup, err := gas.GitHubActions.GetRepoCommitWorkflowSetup(ctx, brokerRequestMessage.Repo,
		brokerRequestMessage.Commit)
	if err != nil {
		gas.App.Logger.Error("could not fetch github workflows setup", "error", err.Error())
		return err
	}
	if repoCommitWorkflowSetup == nil {
		gas.App.Logger.Warn("repo has no github workflows setup")
	}
	workflowsResult, err := gas.waitRepoCommitWorkflows(ctx, repoCommitWorkflowSetup, brokerRequestMessage)
	if err != nil {
		gas.App.Logger.Error("repo has no github workflows setup")
		return err
	}
	resultResponse := broker.ResponseMessage{
		Response: app.BrokerResponseFinished,
		Result:   app.BrokerResultSuccess,
	}
	for _, workflowResult := range workflowsResult {
		resultResponse.ResultDetails = append(resultResponse.ResultDetails, broker.WorkflowDetails{
			WorkflowID:     workflowResult.WorkflowID,
			WorkflowName:   workflowResult.WorkflowName,
			WorkflowResult: workflowResult.Result,
		})
		if workflowResult.Result != githubops.WorkflowResultSuccess {
			resultResponse.Result = app.BrokerResultFailure
		}
	}
	if brokerRequestMessage.PatchEvent != nil {
		err = gas.commentOnPatch(ctx, brokerRequestMessage, resultResponse, repoCommitWorkflowSetup)
		if err != nil {
			gas.App.Logger.Warn("could not comment on patch", "error", err.Error())
		}
	}
	gas.App.Logger.Debug("sending message", "message", fmt.Sprintf("%+v", resultResponse))
	err = gas.Broker.ServeResponse(ctx, resultResponse)
	if err != nil {
		gas.App.Logger.Error(err.Error())
		return err
	}
	return nil
}

func (gas *GitHubActionsServer) commentOnPatch(ctx context.Context, brokerRequestMessage *broker.RequestMessage,
	resultResponse broker.ResponseMessage, gitHubActionsSettings *app.GitHubActionsSettings) error {
	if gitHubActionsSettings == nil {
		return nil
	}
	if len(brokerRequestMessage.PatchEvent.Patch.Revisions) == 0 {
		gas.App.Logger.Warn("could not comment on patch", "error", "no revision found in patch")
		return errors.New("no revision found in patch")
	}
	revision := brokerRequestMessage.PatchEvent.Patch.Revisions[len(brokerRequestMessage.PatchEvent.Patch.
		Revisions)-1]
	commentMessage := gas.preparePatchCommentMessage(resultResponse, *gitHubActionsSettings)
	return gas.Radicle.Comment(ctx, brokerRequestMessage.Repo, brokerRequestMessage.PatchEvent.Patch.ID, revision.ID,
		commentMessage)
}

// preparePatchCommentMessage prepares a message for adding as patch comment with the workflow results.
func (gas *GitHubActionsServer) preparePatchCommentMessage(resultResponse broker.ResponseMessage, gitHubActionsSettings app.GitHubActionsSettings) string {
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

// waitRepoCommitWorkflows waits for all workflows to complete execution and returns their results.
// Wait time is upper bounded by WorkflowCheckInterval.
func (gas *GitHubActionsServer) waitRepoCommitWorkflows(ctx context.Context,
	repoCommitWorkflowSetup *app.GitHubActionsSettings, brokerRequestMessage *broker.RequestMessage) ([]app.
	WorkflowResult, error) {
	if repoCommitWorkflowSetup == nil {
		return nil, nil
	}
	var workflowsResult []app.WorkflowResult
	var err error
	var waitDuration time.Duration
	for {
		workflowsCompleted := true
		workflowsResult, err = gas.GitHubActions.GetRepoCommitWorkflowsResults(ctx, repoCommitWorkflowSetup.GitHubUsername,
			repoCommitWorkflowSetup.GitHubRepo, brokerRequestMessage.Commit)
		if err != nil {
			gas.App.Logger.Error("could not get repo commit workflows", "error", err.Error())
			return nil, err
		}
		for _, workflowResult := range workflowsResult {
			if workflowResult.Status != githubops.WorkflowStatusCompleted {
				workflowsCompleted = false
				break
			}
		}
		if !workflowsCompleted {
			if waitDuration >= time.Second*time.Duration(gas.App.Config.WorkflowsPollTimoutMins) {
				gas.App.Logger.Warn("reached timeout while waiting for workflows to complete")
				break
			}
			time.Sleep(app.WorkflowCheckInterval)
			waitDuration += app.WorkflowCheckInterval
			continue
		}
		break
	}
	return workflowsResult, nil
}
