package radiclegithubactions

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app"
	"radicle-github-actions-adapter/app/githubops"
	"radicle-github-actions-adapter/app/gitops"
	"reflect"
	"strings"
	"testing"
)

type MockGitOps struct{}

func (mgo *MockGitOps) CloneRepoCommit(url, commitHash, repoPath string) error {
	if !strings.Contains(url, "project_id") || repoPath != "/tmp/some_repo_path" || commitHash != "commit_id" {
		return errors.New("invalid params")
	}
	return nil
}

type MockGitHubOps struct{}

func (mgho *MockGitHubOps) CheckRepoCommit(ctx context.Context, user, repo, commit string) error {
	if user != "gh_username" || repo != "gh_reponame" {
		return errors.New("invalid params")
	}
	return nil
}

func (mgho *MockGitHubOps) GetRepoCommitWorkflows(ctx context.Context, user, repo,
	commit string) ([]githubops.WorkflowResult, error) {
	if user != "gh_username" || repo != "gh_reponame" || commit != "commit_id" {
		return nil, errors.New("invalid params")
	}
	result := []githubops.WorkflowResult{
		{
			WorkflowID:   "work_1",
			WorkflowName: "work name1",
			Status:       "completed",
			Result:       "failure",
		},
		{
			WorkflowID:   "work_2",
			WorkflowName: "work name2",
			Status:       "completed",
			Result:       "successful",
		},
	}
	return result, nil
}

func TestRadicleGitHubActions_GetRepoCommitWorkflowSetup(t *testing.T) {
	mockGitOps := MockGitOps{}
	mockGitHubOps := MockGitHubOps{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	ctx := context.WithValue(context.Background(), app.RepoClonePath, "/tmp/some_repo_path")
	type fields struct {
		logger      *slog.Logger
		radicleHome string
		git         gitops.GitOps
		github      githubops.GitHubOps
	}
	type args struct {
		ctx        context.Context
		projectID  string
		commitHash string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		prepareFunc func() error
		want        *app.GitHubActionsSettings
		wantErr     bool
	}{
		{
			name: "GetRepoCommitWorkflowSetup returns existing commit workflow setup",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			prepareFunc: func() error {
				err := os.MkdirAll("/tmp/some_repo_path/.github/workflows", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = os.Create("/tmp/some_repo_path/.github/workflows/some_workflow.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}

				err = os.MkdirAll("/tmp/some_repo_path/.radicle", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				f, err := os.Create("/tmp/some_repo_path/.radicle/github_actions.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = f.WriteString("github_username: gh_username\ngithub_repo: gh_reponame")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				return nil
			},
			args: args{
				ctx:        ctx,
				projectID:  "project_id",
				commitHash: "commit_id",
			},
			want: &app.GitHubActionsSettings{
				GitHubUsername: "gh_username",
				GitHubRepo:     "gh_reponame",
			},
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflowSetup returns nothing when wo github workflows exist",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			prepareFunc: func() error {
				err := os.MkdirAll("/tmp/some_repo_path/.radicle", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				f, err := os.Create("/tmp/some_repo_path/.radicle/github_actions.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = f.WriteString("github_username: gh_username\ngithub_repo: gh_reponame")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				return nil
			},
			args: args{
				ctx:        ctx,
				projectID:  "project_id",
				commitHash: "commit_id",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflowSetup returns nothing when no radicle github actions setup is found",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			prepareFunc: func() error {
				err := os.MkdirAll("/tmp/some_repo_path/.github/workflows", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = os.Create("/tmp/some_repo_path/.github/workflows/some_workflow.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				return nil
			},
			args: args{
				ctx:        ctx,
				projectID:  "project_id",
				commitHash: "commit_id",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflowSetup fails when invalid commit is used",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			prepareFunc: func() error {
				err := os.MkdirAll("/tmp/some_repo_path/.github/workflows", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = os.Create("/tmp/some_repo_path/.github/workflows/some_workflow.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}

				err = os.MkdirAll("/tmp/some_repo_path/.radicle", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				f, err := os.Create("/tmp/some_repo_path/.radicle/github_actions.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = f.WriteString("github_username: gh_username\ngithub_repo: gh_reponame")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				return nil
			},
			args: args{
				ctx:        ctx,
				projectID:  "project_id",
				commitHash: "invalid_commit_id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetRepoCommitWorkflowSetup returns nothing when fail to parse radicle github actions setup",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			prepareFunc: func() error {
				err := os.MkdirAll("/tmp/some_repo_path/.github/workflows", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = os.Create("/tmp/some_repo_path/.github/workflows/some_workflow.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}

				err = os.MkdirAll("/tmp/some_repo_path/.radicle", 0700)
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				f, err := os.Create("/tmp/some_repo_path/.radicle/github_actions.yaml")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				_, err = f.WriteString("SOME INVALID DATA HERE!")
				if err != nil {
					t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
					return err
				}
				return nil
			},
			args: args{
				ctx:        ctx,
				projectID:  "project_id",
				commitHash: "commit_id",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rga := &RadicleGitHubActions{
				logger:      tt.fields.logger,
				radicleHome: tt.fields.radicleHome,
				git:         tt.fields.git,
				github:      tt.fields.github,
			}
			defer os.RemoveAll("/tmp/some_repo_path")
			err := tt.prepareFunc()
			if err != nil {
				t.Errorf("GetRepoCommitWorkflowSetup() could not prepare test, error = %v", err)
				return
			}

			got, err := rga.GetRepoCommitWorkflowSetup(tt.args.ctx, tt.args.projectID, tt.args.commitHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoCommitWorkflowSetup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRepoCommitWorkflowSetup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRadicleGitHubActions_GetRepoCommitWorkflowsResults(t *testing.T) {
	mockGitOps := MockGitOps{}
	mockGitHubOps := MockGitHubOps{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	ctx := context.WithValue(context.Background(), app.RepoClonePath, "/tmp/random_repo_path")
	type fields struct {
		logger      *slog.Logger
		radicleHome string
		git         gitops.GitOps
		github      githubops.GitHubOps
	}
	type args struct {
		ctx            context.Context
		githubUsername string
		githubRepo     string
		githubCommit   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []app.WorkflowResult
		wantErr bool
	}{
		{
			name: "GetRepoCommitWorkflowsResults returns all results",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			args: args{
				ctx:            ctx,
				githubUsername: "gh_username",
				githubRepo:     "gh_reponame",
				githubCommit:   "commit_id",
			},
			want: []app.WorkflowResult{
				{
					WorkflowID:   "work_1",
					WorkflowName: "work name1",
					Status:       "completed",
					Result:       "failure",
				},
				{
					WorkflowID:   "work_2",
					WorkflowName: "work name2",
					Status:       "completed",
					Result:       "successful",
				},
			},
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflowsResults fails when invalid repo is used",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			args: args{
				ctx:            ctx,
				githubUsername: "gh_username",
				githubRepo:     "INVALID_REPO_NAME",
				githubCommit:   "commit_id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetRepoCommitWorkflowsResults fails when invalid commit is used",
			fields: fields{
				logger:      logger,
				radicleHome: "/home/user",
				git:         &mockGitOps,
				github:      &mockGitHubOps,
			},
			args: args{
				ctx:            ctx,
				githubUsername: "gh_username",
				githubRepo:     "gh_reponame",
				githubCommit:   "INVALID_COMMIT_ID",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rga := &RadicleGitHubActions{
				logger:      tt.fields.logger,
				radicleHome: tt.fields.radicleHome,
				git:         tt.fields.git,
				github:      tt.fields.github,
			}
			got, err := rga.GetRepoCommitWorkflowsResults(tt.args.ctx, tt.args.githubUsername, tt.args.githubRepo, tt.args.githubCommit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoCommitWorkflowsResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRepoCommitWorkflowsResults() got = %v, want %v", got, tt.want)
			}
		})
	}
}
