package serve

import (
	"context"
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
	WorkflowsStartLagSecs   uint64
	WorkflowsPollTimoutSecs uint64
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
		gas.App.Logger.Error("could not parse broker request message", "error", err.Error())
		return err
	}
	gas.App.Logger.Debug("serving broker message", "message", *brokerRequestMessage)
	jobResponse := broker.ResponseMessage{
		Response: app.BrokerResponseTriggered,
		RunID: &broker.RunID{
			ID: eventUUID,
		},
	}
	gas.App.Logger.Debug("sending message", "message", jobResponse)
	err = gas.Broker.ServeResponse(ctx, jobResponse)
	if err != nil {
		gas.App.Logger.Error("could not send response message to broker", "error", err.Error())
		return err
	}
	repoCommitWorkflowSetup, err := gas.GitHubActions.GetRepoCommitWorkflowSetup(ctx, brokerRequestMessage.Repo,
		brokerRequestMessage.Commit)
	if err != nil {
		gas.App.Logger.Error("could not fetch github workflows setup", "error", err.Error())
		return err
	}
	resultResponse := broker.ResponseMessage{
		Response: app.BrokerResponseFinished,
		Result:   app.BrokerResultSuccess,
	}
	if repoCommitWorkflowSetup != nil {
		// Write 1st comment that we check GitHub for workflows
		if brokerRequestMessage.PatchEvent != nil {
			commentMessage := "Checking for GitHub Actions Workflows..."
			_ = gas.commentOnPatch(ctx, brokerRequestMessage, commentMessage)
		}
		time.Sleep(time.Second * time.Duration(gas.App.Config.WorkflowsStartLagSecs))

		//Wait for GitHub Workflows results and write comment
		workflowsResult, err := gas.waitRepoCommitWorkflows(ctx, repoCommitWorkflowSetup, brokerRequestMessage)
		if err != nil {
			gas.App.Logger.Error("failed waiting for github workflows")
			return err
		}
		gas.updateResponseResults(&resultResponse, workflowsResult)
		if brokerRequestMessage.PatchEvent != nil {
			commentMessage := gas.preparePatchCommentResultMessage(resultResponse, *repoCommitWorkflowSetup)
			_ = gas.commentOnPatch(ctx, brokerRequestMessage, commentMessage)
		}
	}
	gas.App.Logger.Debug("sending message", "message", resultResponse)
	err = gas.Broker.ServeResponse(ctx, resultResponse)
	if err != nil {
		gas.App.Logger.Error("could not send response message to broker", "error", err.Error())
		return err
	}
	return nil
}

// updateResponseResults
func (gas *GitHubActionsServer) updateResponseResults(resultResponse *broker.ResponseMessage, workflowsResult []app.
	WorkflowResult) {
	for _, workflowResult := range workflowsResult {
		workflowDetails := broker.WorkflowDetails{
			WorkflowID:     workflowResult.WorkflowID,
			WorkflowName:   workflowResult.WorkflowName,
			WorkflowResult: workflowResult.Result,
		}
		if len(workflowDetails.WorkflowResult) == 0 {
			workflowDetails.WorkflowResult = workflowResult.Status
		}
		resultResponse.ResultDetails = append(resultResponse.ResultDetails, workflowDetails)
		if resultResponse.Response == app.BrokerResponseFinished &&
			workflowResult.Result != githubops.WorkflowResultSuccess {
			resultResponse.Result = app.BrokerResultFailure
		}
	}
}

// waitRepoCommitWorkflows waits for all workflows to complete execution and returns their results.
// Wait time is upper bounded by WorkflowCheckInterval.
func (gas *GitHubActionsServer) waitRepoCommitWorkflows(ctx context.Context,
	repoCommitWorkflowSetup *app.GitHubActionsSettings, brokerRequestMessage *broker.RequestMessage) ([]app.
	WorkflowResult, error) {
	var workflowsResult []app.WorkflowResult
	var err error
	for start := time.Now(); time.Since(start) < time.Second*time.Duration(gas.App.Config.WorkflowsPollTimoutSecs); {
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
		if workflowsCompleted {
			gas.App.Logger.Info("all workflows execution complete")
			break
		}
		if brokerRequestMessage.PatchEvent != nil {
			resultResponse := broker.ResponseMessage{
				Response: app.BrokerResponseInProgress,
				Result:   app.BrokerResultSuccess,
			}
			gas.updateResponseResults(&resultResponse, workflowsResult)
			commentMessage := gas.preparePatchCommentResultMessage(resultResponse, *repoCommitWorkflowSetup)
			_ = gas.commentOnPatch(ctx, brokerRequestMessage, commentMessage)
		}
		time.Sleep(app.WorkflowCheckInterval)
	}
	return workflowsResult, nil
}
