package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v57/github"
	"log/slog"
	"os"
	"radicle-github-actions-adapter/app/githubops"
	"reflect"
	"strconv"
	"testing"
)

type MockGitHub struct {
	repos   Repos
	actions Actions
}

type Repos struct{}
type Actions struct{}

func (r *Repos) GetCommit(ctx context.Context, owner, repo, sha string,
	opts *github.ListOptions) (*github.RepositoryCommit,
	*github.Response, error) {
	if owner == "repo_owner" && repo == "repo_name" && sha == "commit_hash" {
		return &github.RepositoryCommit{}, &github.Response{}, nil
	}
	return nil, nil, errors.New("an error occurred")
}

func (a *Actions) ListRepositoryWorkflowRuns(ctx context.Context, owner, repo string,
	opts *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	// in order to mock this function we will use the repo to return the appropriate amount of workflows
	if owner == "repo_owner" {
		totalWorkflows, err := strconv.Atoi(repo)
		if err != nil {
			return nil, nil, errors.New("invalid repos")
		}

		if totalWorkflows < (opts.Page)*opts.PerPage {
			return nil, nil, errors.New(fmt.Sprintf("request exceeds total workflows %v, %v, %v", totalWorkflows,
				opts.Page, opts.PerPage))
		}
		runs := github.WorkflowRuns{
			TotalCount:   &totalWorkflows,
			WorkflowRuns: []*github.WorkflowRun{},
		}
		for i := 0; i < totalWorkflows; i++ {
			workId := int64(i)
			workName := "work " + strconv.Itoa(i)
			statusCompleted := githubops.WorkflowStatusCompleted
			result := githubops.WorkflowResultSuccess
			if i%2 == 0 {
				result = githubops.WorkflowResultFailure
			}
			runs.WorkflowRuns = append(runs.WorkflowRuns, &github.WorkflowRun{
				ID:         &workId,
				Name:       &workName,
				Status:     &statusCompleted,
				Conclusion: &result,
			})
		}
		return &runs, &github.Response{}, nil
	}
	return nil, nil, errors.New("an error occurred")
}

func TestGitHub_CheckRepoCommit(t *testing.T) {
	mGH := MockGitHub{}
	type fields struct {
		logger  *slog.Logger
		pat     string
		repos   RepositoriesService
		actions ActionsService
	}
	type args struct {
		ctx    context.Context
		user   string
		repo   string
		commit string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "CheckRepoCommit is successful",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "repo_owner",
				repo:   "repo_name",
				commit: "commit_hash",
			},
			wantErr: false,
		},
		{
			name: "CheckRepoCommit fails when CheckRepoCommit fails ",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "unknown_owner",
				repo:   "repo_name",
				commit: "commit_hash",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := &GitHub{
				logger:  tt.fields.logger,
				pat:     tt.fields.pat,
				repos:   tt.fields.repos,
				actions: tt.fields.actions,
			}
			if err := gh.CheckRepoCommit(tt.args.ctx, tt.args.user, tt.args.repo, tt.args.commit); (err != nil) != tt.wantErr {
				t.Errorf("CheckRepoCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHub_GetRepoCommitWorkflows(t *testing.T) {
	mGH := MockGitHub{}
	type fields struct {
		logger  *slog.Logger
		pat     string
		repos   RepositoriesService
		actions ActionsService
	}
	type args struct {
		ctx    context.Context
		user   string
		repo   string
		commit string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []githubops.WorkflowResult
		wantErr bool
	}{
		{
			name: "GetRepoCommitWorkflows is successful with 1 workflow",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "repo_owner",
				repo:   "1",
				commit: "commit_hash",
			},
			want: []githubops.WorkflowResult{
				{
					WorkflowID:   "0",
					WorkflowName: "work 0",
					Status:       githubops.WorkflowStatusCompleted,
					Result:       githubops.WorkflowResultFailure,
				},
			},
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflows is successful with 0 workflow",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "repo_owner",
				repo:   "0",
				commit: "commit_hash",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflows is successful with lots workflows",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "repo_owner",
				repo:   "100",
				commit: "commit_hash",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "GetRepoCommitWorkflows fails with invalid user",
			fields: fields{
				logger:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
				pat:     "github_pat",
				repos:   &mGH.repos,
				actions: &mGH.actions,
			},
			args: args{
				ctx:    context.Background(),
				user:   "invalid_owner",
				repo:   "0",
				commit: "commit_hash",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := &GitHub{
				logger:  tt.fields.logger,
				pat:     tt.fields.pat,
				repos:   tt.fields.repos,
				actions: tt.fields.actions,
			}
			got, err := gh.GetRepoCommitWorkflows(tt.args.ctx, tt.args.user, tt.args.repo, tt.args.commit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoCommitWorkflows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strconv.Itoa(len(got)) != tt.args.repo {
				t.Errorf("GetRepoCommitWorkflows() expected  %v workflows, got %v", tt.args.repo, strconv.Itoa(len(got)))
			}
			if tt.args.repo != "100" {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetRepoCommitWorkflows() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNewGitHub(t *testing.T) {
	type args struct {
		pat    string
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *GitHub
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGitHub(tt.args.pat, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGitHub() = %v, want %v", got, tt.want)
			}
		})
	}
}
