package serve

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/broker"
	"radicle-github-actions-adapter/app/githubops"
	"radicle-github-actions-adapter/app/radicle"
	"strconv"
	"strings"
	"testing"
)

type MockBroker struct{}

func (mb *MockBroker) ParseRequestMessage(ctx context.Context) (*broker.RequestMessage, error) {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	commitID := eventUUID[len(eventUUID)-1:]
	if strings.HasPrefix(eventUUID, "event-uuid-push") {
		brokerMessage := broker.RequestMessage{
			Repo:   "repo_id",
			Commit: commitID,
			PushEvent: &broker.RequestPushEventMessage{
				Request:   "trigger",
				EventType: "push",
				Pusher: broker.Pusher{
					ID:    "alias_id",
					Alias: "alias_name",
				},
				Before:  "before_commit_hash",
				After:   commitID,
				Commits: []string{"before_commit_hash", "1"},
				Repository: broker.Repository{
					ID:            "repo_id",
					Name:          "repo name",
					Description:   "",
					Private:       false,
					DefaultBranch: "main",
					Delegates:     nil,
				},
			},
			PatchEvent: nil,
		}
		return &brokerMessage, nil
	} else if strings.HasPrefix(eventUUID, "event-uuid-patch") {
		brokerMessage := broker.RequestMessage{
			Repo:      "repo_id",
			Commit:    commitID,
			PushEvent: nil,
			PatchEvent: &broker.RequestPatchEventMessage{
				Request:   "trigger",
				EventType: "patch",
				Action:    "created",
				Patch: broker.PatchDetails{
					ID: "patch_id",
					Author: broker.Author{
						ID:    "alias_id",
						Alias: "alias_name",
					},
					Title: "patch title",
					State: broker.PatchState{
						Status:    "open",
						Conflicts: nil,
					},
					Before:    "before_commit_hash",
					After:     commitID,
					Commits:   []string{"before_commit_hash", "1"},
					Target:    "delegates",
					Labels:    nil,
					Assignees: nil,
					Revisions: []broker.PatchRevision{
						{
							ID: "revision_id",
							Author: broker.Author{
								ID:    "revision_author_id",
								Alias: "revision_author_name",
							},
							Description: "",
							Base:        "",
							Oid:         "",
							Timestamp:   0,
						},
					},
				},
				Repository: broker.Repository{
					ID:            "repo_id",
					Name:          "repo name",
					Description:   "",
					Private:       false,
					DefaultBranch: "main",
					Delegates:     nil,
				},
			},
		}
		return &brokerMessage, nil
	} else if strings.HasPrefix(eventUUID, "event-uuid-invalid-push") {
		brokerMessage := broker.RequestMessage{
			Repo:   "repo_id",
			Commit: "after_commit_hash",
			PushEvent: &broker.RequestPushEventMessage{
				Request:   "invalid",
				EventType: "push",
			},
			PatchEvent: nil,
		}
		return &brokerMessage, nil
	} else if strings.HasPrefix(eventUUID, "event-uuid-invalid-patch") {
		brokerMessage := broker.RequestMessage{
			Repo:      "repo_id",
			Commit:    "after_commit_hash",
			PushEvent: nil,
			PatchEvent: &broker.RequestPatchEventMessage{
				Request:   "invalid",
				EventType: "patch",
				Action:    "created",
			},
		}
		return &brokerMessage, nil
	}
	return nil, errors.New("unknown error")
}

func (mb *MockBroker) ServeResponse(ctx context.Context, responseMessage broker.ResponseMessage) error {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	if strings.Contains(eventUUID, "invalid") {
		return errors.New("unknown error")
	}
	return nil
}

func (mb *MockBroker) ServeErrorResponse(ctx context.Context, responseErrorMessage broker.ResponseErrorMessage) error {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	if strings.Contains(eventUUID, "invalid") {
		return errors.New("unknown error")
	}
	return nil
}

type MockGitHubActions struct{}

func (g *MockGitHubActions) GetRepoCommitWorkflowSetup(ctx context.Context, projectID, commitHash string) (*app.GitHubActionsSettings, error) {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	if strings.Contains(eventUUID, "invalid") {
		return nil, errors.New("unknown error")
	}
	return &app.GitHubActionsSettings{
		GitHubUsername: "repo_user",
		GitHubRepo:     "repo_name",
	}, nil
}

func (g *MockGitHubActions) GetRepoCommitWorkflowsResults(ctx context.Context, githubUsername, githubRepo, githubCommit string) ([]app.WorkflowResult, error) {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	if strings.Contains(eventUUID, "invalid") {
		return nil, errors.New("unknown error")
	}
	if githubUsername != "repo_user" || githubRepo != "repo_name" {
		return nil, errors.New("invalid data")
	}
	totalWorkflows, err := strconv.Atoi(githubCommit)
	if err != nil {
		return nil, errors.New("invalid repos")
	}

	workflowResults := []app.WorkflowResult{}
	for i := 0; i < totalWorkflows; i++ {
		workId := i
		result := githubops.WorkflowResultSuccess
		if i%2 == 0 {
			result = githubops.WorkflowResultFailure
		}
		workflowResults = append(workflowResults, app.WorkflowResult{
			WorkflowID:   strconv.Itoa(workId),
			WorkflowName: "work " + strconv.Itoa(i),
			Status:       githubops.WorkflowStatusCompleted,
			Result:       result,
		})
	}
	return workflowResults, nil
}

type MockRadiclePatch struct{}

func (p *MockRadiclePatch) Comment(ctx context.Context, repoID, patchID, revisionID string, message string) error {
	eventUUID := ctx.Value(app.EventUUIDKey).(string)
	if strings.Contains(eventUUID, "invalid") {
		return errors.New("unknown error")
	}
	if repoID != "repo_id" {
		return errors.New("invalid data")
	}
	totalWorkflowsString := eventUUID[len(eventUUID)-1:]
	totalWorkflows, err := strconv.Atoi(totalWorkflowsString)
	if err != nil {
		return err
	}

	if totalWorkflows != strings.Count(message, "<a href") {
		return errors.New("total workflows do not match message")
	}

	return nil
}

func TestGitHubActions_Serve(t *testing.T) {
	mockBroker := MockBroker{}
	mockGitHubActions := MockGitHubActions{}
	mockRadicle := MockRadiclePatch{}
	type fields struct {
		App           *App
		Broker        broker.Broker
		GitHubActions app.GitHubActions
		Radicle       radicle.Patch
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Serve is successful with single workflow on push event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-push-valid-1"), app.RepoClonePathKey, "event-uuid-push-valid-1"),
			},
			wantErr: false,
		},
		{
			name: "Serve is successful with two workflows on push event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-push-valid-2"), app.RepoClonePathKey, "event-uuid-push-valid-2"),
			},
			wantErr: false,
		},
		{
			name: "Serve fails with invalid push event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-push-invalid-1"), app.RepoClonePathKey, "event-uuid-push-invalid-1"),
			},
			wantErr: true,
		},

		{
			name: "Serve fails when broker message is invalid",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-invalid"), app.RepoClonePathKey, "event-uuid-invalid"),
			},
			wantErr: true,
		},
		{
			name: "Serve is successful with single workflow on patch event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-patch-valid-1"), app.RepoClonePathKey, "event-uuid-patch-valid-1"),
			},
			wantErr: false,
		},
		{
			name: "Serve is successful with two workflows on patch event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-patch-valid-2"), app.RepoClonePathKey, "event-uuid-patch-valid-2"),
			},
			wantErr: false,
		},
		{
			name: "Serve fails with invalid patch event",
			fields: fields{
				App: &App{
					Config: AppConfig{
						RadicleHome:             ".radicle/",
						GitHubPAT:               "github_path",
						WorkflowsPollTimoutSecs: 10,
						RadicleHttpdURL:         "http://radicle.url",
						RadicleSessionToken:     "rad_session_id",
					},
					Logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				},
				Broker:        &mockBroker,
				GitHubActions: &mockGitHubActions,
				Radicle:       &mockRadicle,
			},
			args: args{
				ctx: context.WithValue(context.WithValue(context.Background(), app.EventUUIDKey,
					"event-uuid-patch-invalid-1"), app.RepoClonePathKey, "event-uuid-patch-invalid-1"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas := &GitHubActionsServer{
				App:           tt.fields.App,
				Broker:        tt.fields.Broker,
				GitHubActions: tt.fields.GitHubActions,
				Radicle:       tt.fields.Radicle,
			}
			if err := gas.Serve(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubActions_PreparePatchCommentMessage(t *testing.T) {
	gas := GitHubActionsServer{}

	githubActionsSettings := app.GitHubActionsSettings{
		GitHubUsername: "testUser",
		GitHubRepo:     "testRepo",
	}

	cases := []struct {
		name     string
		response broker.ResponseMessage
		expected string
	}{
		{
			name: "PreparePatchCommentMessage is successful using only successful results",
			response: broker.ResponseMessage{
				Result: githubops.WorkflowResultSuccess,
				ResultDetails: []broker.WorkflowDetails{
					{WorkflowID: "1", WorkflowName: "BuildTest", WorkflowResult: githubops.WorkflowResultSuccess},
					{WorkflowID: "2", WorkflowName: "UnitTests", WorkflowResult: githubops.WorkflowResultFailure},
				},
			},
			expected: "Github Actions Result: success ✅\n\nDetails:\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/1\" target=\"_blank\" >BuildTest (1)</a>: success 🟢\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/2\" target=\"_blank\" >UnitTests (2)</a>: failure 🔴",
		},
		{
			name: "PreparePatchCommentMessage is successful using only failed results",
			response: broker.ResponseMessage{
				Result: githubops.WorkflowResultFailure,
				ResultDetails: []broker.WorkflowDetails{
					{WorkflowID: "1", WorkflowName: "BuildTest", WorkflowResult: githubops.WorkflowResultSuccess},
					{WorkflowID: "2", WorkflowName: "UnitTests", WorkflowResult: githubops.WorkflowResultFailure},
				},
			},
			expected: "Github Actions Result: failure ❌\n\nDetails:\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/1\" target=\"_blank\" >BuildTest (1)</a>: success 🟢\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/2\" target=\"_blank\" >UnitTests (2)</a>: failure 🔴",
		},
		{
			name: "PreparePatchCommentMessage is successful using mixed results",
			response: broker.ResponseMessage{
				Result: githubops.WorkflowResultFailure,
				ResultDetails: []broker.WorkflowDetails{
					{WorkflowID: "1", WorkflowName: "BuildTest", WorkflowResult: githubops.WorkflowResultSuccess},
					{WorkflowID: "2", WorkflowName: "UnitTests", WorkflowResult: githubops.WorkflowResultFailure},
					{WorkflowID: "3", WorkflowName: "IntegrationTests", WorkflowResult: "otherResult"},
				},
			},
			expected: "Github Actions Result: failure ❌\n\nDetails:\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/1\" target=\"_blank\" >BuildTest (1)</a>: success 🟢\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/2\" target=\"_blank\" >UnitTests (2)</a>: failure 🔴\n\n - <a href=\"https://github.com/testUser/testRepo/actions/runs/3\" target=\"_blank\" >IntegrationTests (3)</a>: otherResult 🟡",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := gas.preparePatchCommentMessage(tc.response, githubActionsSettings)
			if result != tc.expected {
				t.Fatalf("expected %s, but got %s", tc.expected, result)
			}
		})
	}
}
