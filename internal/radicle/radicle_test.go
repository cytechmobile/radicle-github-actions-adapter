package radicle

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"radicle-github-actions-adapter/app"
	"strings"
	"testing"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

func TestRadicle_Comment(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	ctx := context.WithValue(context.Background(), app.RepoClonePathKey, "repo_path")
	type fields struct {
		nodeURL      string
		token        string
		clientDoFunc func(req *http.Request) (*http.Response, error)
		logger       *slog.Logger
	}
	type args struct {
		ctx        context.Context
		repoID     string
		patchID    string
		revisionID string
		message    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Comment adds a comment to the given patch",
			fields: fields{
				nodeURL: "http://node.url",
				token:   "some_token",
				logger:  logger,
				clientDoFunc: func(req *http.Request) (*http.Response, error) {
					if "http://node.url/api/v1/projects/repo_id/patches/patch_id" != req.URL.String() {
						t.Errorf("Add patch comment request URL got = %v, want %v", req.URL.String(),
							"http://node.url/api/v1/projects/repo_id/patches/patch_id")
					}
					if req.Method != http.MethodPatch {
						t.Errorf("Add patch comment request Method got = %v, want %v", req.Method, http.MethodPatch)
					}
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Errorf("Add patch comment  could not read request body %v", err)
					}
					if !strings.Contains(string(body), "some very important message") {
						t.Errorf("Add patch comment request payload got = %v, want %v", string(body), "some very important message")
					}
					if req.Header.Get("content-type") != "application/json" {
						t.Errorf("Add patch comment request header content-type got = %v, want %v",
							req.Header.Get("content-type"), "application/json")
					}

					if req.Header.Get("Authorization") != "Bearer some_token" {
						t.Errorf("Add patch comment request header Authorization got = %v, want %v",
							req.Header.Get("Authorization"), "Bearer some_token")
					}
					resp := http.Response{
						StatusCode: http.StatusOK,
						Body:       http.NoBody,
					}
					return &resp, nil
				},
			},
			args: args{
				ctx:        ctx,
				repoID:     "repo_id",
				patchID:    "patch_id",
				revisionID: "revision_id",
				message:    "some very important message",
			},
			wantErr: false,
		},
		{
			name: "Comment fails to add comment to the given patch returns error",
			fields: fields{
				nodeURL: "http://node.url",
				token:   "some_token",
				logger:  logger,
				clientDoFunc: func(req *http.Request) (*http.Response, error) {
					if "http://node.url/api/v1/projects/repo_id/patches/patch_id" != req.URL.String() {
						t.Errorf("Add patch comment request URL got = %v, want %v", req.URL.String(),
							"http://node.url/api/v1/projects/repo_id/patches/patch_id")
					}
					if req.Method != http.MethodPatch {
						t.Errorf("Add patch comment request Method got = %v, want %v", req.Method, http.MethodPatch)
					}
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Errorf("Add patch comment  could not read request body %v", err)
					}
					if !strings.Contains(string(body), "some very important message") {
						t.Errorf("Add patch comment request payload got = %v, want %v", string(body), "some very important message")
					}
					if req.Header.Get("content-type") != "application/json" {
						t.Errorf("Add patch comment request header content-type got = %v, want %v",
							req.Header.Get("content-type"), "application/json")
					}

					if req.Header.Get("Authorization") != "Bearer some_token" {
						t.Errorf("Add patch comment request header Authorization got = %v, want %v",
							req.Header.Get("Authorization"), "Bearer some_token")
					}
					resp := http.Response{
						StatusCode: http.StatusUnprocessableEntity,
						Body:       http.NoBody,
					}
					return &resp, nil
				},
			},
			args: args{
				ctx:        ctx,
				repoID:     "repo_id",
				patchID:    "patch_id",
				revisionID: "revision_id",
				message:    "some very important message",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := MockHTTPClient{
				DoFunc: tt.fields.clientDoFunc,
			}
			r := &Radicle{
				nodeURL: tt.fields.nodeURL,
				token:   tt.fields.token,
				client:  &mockClient,
				logger:  tt.fields.logger,
			}
			if err := r.Comment(tt.args.ctx, tt.args.repoID, tt.args.patchID, tt.args.revisionID, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Comment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
