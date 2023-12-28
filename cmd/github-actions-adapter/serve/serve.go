package serve

import (
	"context"
	"fmt"
	"log/slog"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
)

type AppConfig struct {
	RadicleHome string
	GitHubPAT   string
}

type App struct {
	Config AppConfig
	Logger *slog.Logger
}

// GitHubActionsServer is used as a container for the most important dependencies.
type GitHubActionsServer struct {
	App    *App
	Broker broker.Broker
}

// NewGitHubActionsServer returns a pointer to a new GitHub Action Server.
func NewGitHubActionsServer(config *App, broker broker.Broker) *GitHubActionsServer {
	server := &GitHubActionsServer{
		App:    config,
		Broker: broker,
	}
	return server
}

// Serve is responsible for parsing stdin input and check any GitHub Actions status of radicle projects
// It also manages replies to the broker.
func (whs *GitHubActionsServer) Serve(ctx context.Context) error {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	whs.App.Logger.Info("serving event", app.EventUUIDKey, eventUUID)
	brokerRequestMessage, err := whs.Broker.ParseRequestMessage(ctx)
	if err != nil {
		whs.App.Logger.Error(err.Error())
		return err
	}
	whs.App.Logger.Debug("received message type", "message", fmt.Sprintf("%+v", *brokerRequestMessage))

	jobResponse := broker.ResponseMessage{
		Response: app.BrokerResponseTriggered,
		RunID: &broker.RunID{
			ID: eventUUID,
		},
	}
	whs.App.Logger.Debug("sending message", "message", fmt.Sprintf("%+v", jobResponse))
	err = whs.Broker.ServeResponse(ctx, jobResponse)
	if err != nil {
		whs.App.Logger.Error(err.Error())
		return err
	}

	// Check GitHub Actions status for repo's commit

	resultResponse := broker.ResponseMessage{
		Response: app.BrokerResponseFinished,
		Result:   app.BrokerResultSuccess,
	}
	whs.App.Logger.Debug("sending message", "message", fmt.Sprintf("%+v", resultResponse))
	err = whs.Broker.ServeResponse(ctx, resultResponse)
	if err != nil {
		whs.App.Logger.Error(err.Error())
		return err
	}

	return nil
}
