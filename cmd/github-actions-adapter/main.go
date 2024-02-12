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
	"radicle-github-actions-adapter/internal/radicle"
	"radicle-github-actions-adapter/internal/radiclegithubactions"
	"radicle-github-actions-adapter/internal/readerwriterbroker"
	"radicle-github-actions-adapter/pkg/env"
	"radicle-github-actions-adapter/pkg/gohome"
	"radicle-github-actions-adapter/pkg/version"
	"strings"
)

func main() {
	envLogLevel := env.GetString("LOG_LEVEL", "info")
	logLevel := flag.String("loglevel", envLogLevel, "Log level: debug, info, warn, error")
	showVersion := flag.Bool("version", false, "Display binary version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		os.Exit(0)
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

	slogLevel := LogLevelToSlogLevel(logLevel)
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
	logger.Info("radicle-github-actions-adapter terminated successfully")
}

func run(logger *slog.Logger) error {
	var cfg serve.AppConfig
	cfg.RadicleHome = gohome.Expand(env.GetString("RAD_HOME", "~/.radicle"))
	cfg.RadicleHttpdURL = env.GetString("RAD_HTTPD_URL", "http://127.0.0.1:8080")
	cfg.RadicleSessionToken = env.GetString("RAD_SESSION_TOKEN", "")
	cfg.GitHubPAT = env.GetString("GITHUB_PAT", "")
	cfg.WorkflowsStartLagSecs = env.GetUint64("WORKFLOWS_START_LAG_SECS", 60)
	if cfg.WorkflowsStartLagSecs == 0 {
		cfg.WorkflowsStartLagSecs = 60
	}
	cfg.WorkflowsPollTimoutSecs = env.GetUint64("WORKFLOWS_POLL_TIMEOUT_SECS", 30*60)
	if cfg.WorkflowsPollTimoutSecs == 0 {
		cfg.WorkflowsPollTimoutSecs = 30 * 60
	}

	logger.Debug("starting with configuration", "RadicleHome", cfg.RadicleHome, "RadicleHttpdURL", cfg.RadicleHttpdURL,
		"RadicleSessionToken", cfg.RadicleSessionToken, "WorkflowsPollTimoutSecs", cfg.WorkflowsPollTimoutSecs,
		"GitHubPAT length", len(cfg.GitHubPAT))

	var application serve.App
	application.Config = cfg
	application.Logger = logger
	logger.Info("radicle-github-actions-adapter is starting", "version", version.Get())
	radicleBroker := readerwriterbroker.NewReaderWriterBroker(os.Stdin, os.Stdout, logger)
	gitOps := git.NewGit(logger)
	gitHubOps := github.NewGitHub(cfg.GitHubPAT, logger)
	gitHubActions := radiclegithubactions.NewRadicleGitHubActions(cfg.RadicleHome, gitOps, gitHubOps, logger)
	radiclePatch := radicle.NewRadicle(cfg.RadicleHttpdURL, cfg.RadicleSessionToken, logger)
	srv := serve.NewGitHubActionsServer(&application, radicleBroker, gitHubActions, radiclePatch)

	eventUUID := uuid.New().String()
	ctx := context.WithValue(context.Background(), app.EventUUIDKey, eventUUID)
	ctx = context.WithValue(ctx, app.RepoClonePathKey, eventUUID)
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%+v", r)
			_ = handleAppError(ctx, logger, err, radicleBroker)
		}
	}()
	err := srv.Serve(ctx)
	if err != nil {
		return handleAppError(ctx, logger, err, radicleBroker)
	}

	return nil
}

func handleAppError(ctx context.Context, logger *slog.Logger, err error,
	radicleBroker *readerwriterbroker.ReaderWriterBroker) error {
	logger.Error("could not serve radicle gitHub actions", "error", err.Error())
	resultErrorResponse := broker.ResponseErrorMessage{
		Response: app.BrokerResponseFinished,
		Result: broker.ErrorMessage{
			Error: err.Error(),
		},
	}
	err = radicleBroker.ServeErrorResponse(ctx, resultErrorResponse)
	if err != nil {
		logger.Error("could not respond to broker", "error", err.Error())
		return err
	}
	return nil
}

func LogLevelToSlogLevel(logLevel *string) slog.Level {
	slogLevel := slog.LevelInfo
	if logLevel != nil {
		switch *logLevel {
		case "debug":
			slogLevel = slog.LevelDebug
		case "info":
			slogLevel = slog.LevelInfo
		case "warn":
			slogLevel = slog.LevelWarn
		case "error":
			slogLevel = slog.LevelError
		}
	}
	return slogLevel
}
