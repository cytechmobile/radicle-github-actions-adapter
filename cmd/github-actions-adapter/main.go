package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
	"radicle-github-actions-adapter/cmd/github-actions-adapter/serve"
	"radicle-github-actions-adapter/internal/git"
	"radicle-github-actions-adapter/internal/github"
	"radicle-github-actions-adapter/internal/radiclegithubactions"
	"radicle-github-actions-adapter/internal/readerwriterbroker"
	"radicle-github-actions-adapter/pkg/env"
	"radicle-github-actions-adapter/pkg/gohome"
	"radicle-github-actions-adapter/pkg/version"
	"strings"
)

func main() {
	logLevel := flag.String("loglevel", "info", "Log slogLevel: debug, info, warn, error")
	showVersion := flag.Bool("version", false, "Display binary version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		os.Exit(0)
	}

	slogLevel := slog.LevelInfo
	if logLevel != nil {
		switch *logLevel {
		case "debug":
			slogLevel = slog.LevelDebug
		case "info":
			slogLevel = slog.LevelInfo
		case "wanr":
			slogLevel = slog.LevelWarn
		case "error":
			slogLevel = slog.LevelError
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to determine working directory")
		os.Exit(1)
	}

	//Remove absolute path from logger
	replacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			if file, ok := strings.CutPrefix(source.File, wd); ok {
				source.File = file
			}
		}
		return a
	}
	slogHandlerOptions := slog.HandlerOptions{Level: slogLevel, ReplaceAttr: replacer}

	//Print source only on debug level
	if slogLevel == slog.LevelDebug {
		slogHandlerOptions.AddSource = true
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slogHandlerOptions))

	slog.SetDefault(logger)
	err = run(logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var cfg serve.AppConfig
	cfg.RadicleHome = gohome.Expand(env.GetString("RAD_HOME", "~/.radicle"))
	cfg.GitHubPAT = env.GetString("GITHUB_PAT", "")
	cfg.WorkflowsPollTimoutSecs = env.GetUint64("WORKFLOWS_POLL_TIMEOUT_SECS", 6000)
	logger.Debug("starting with configuration", "RadicleHome", cfg.RadicleHome,
		"WorkflowsPollTimoutSecs", cfg.WorkflowsPollTimoutSecs, "GitHubPAT length", len(cfg.GitHubPAT))

	var application serve.App
	application.Config = cfg
	application.Logger = logger
	logger.Info("radicle-github-actions-adapter is starting", "version", version.Get())
	radicleBroker := readerwriterbroker.NewReaderWriterBroker(os.Stdin, os.Stdout, logger)
	gitOps := git.NewGit(logger)
	gitHubOps := github.NewGitHub(cfg.GitHubPAT, logger)
	gitHubActions := radiclegithubactions.NewRadicleGitHubActions(cfg.RadicleHome, gitOps, gitHubOps, logger)
	srv := serve.NewGitHubActionsServer(&application, radicleBroker, gitHubActions)

	eventUUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), app.EventUUIDKey, eventUUID)
	ctx = context.WithValue(ctx, app.RepoClonePath, eventUUID)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("could not serve radicle GitHub Actions", "error", r)
			resultErrorResponse := broker.ResponseErrorMessage{
				Response: app.BrokerResponseFinished,
				Result: broker.ErrorMessage{
					Error: fmt.Sprintf("%+v", r),
				},
			}
			radicleBroker.ServeErrorResponse(ctx, resultErrorResponse)
		}
	}()

	err := srv.Serve(ctx)
	if err != nil {
		resultErrorResponse := broker.ResponseErrorMessage{
			Response: app.BrokerResponseFinished,
			Result: broker.ErrorMessage{
				Error: err.Error(),
			},
		}
		radicleBroker.ServeErrorResponse(ctx, resultErrorResponse)
		logger.Error("could not serve radicle GitHub Actions", "error", err.Error())
	}
	return err
}
