package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestGit_CloneRepoCommit(t *testing.T) {
	type fields struct {
		logger *slog.Logger
	}
	type args struct {
		url        string
		commitHash string
		repoPath   string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		prepareFunc func() (string, error)
		cleanupFunc func()
		wantErr     bool
	}{
		{
			name: "CloneRepoCommit is successful",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
			},
			args: args{
				url:        "file:///tmp/repo_name",
				commitHash: "",
				repoPath:   "/tmp/cloned_repo_name",
			},
			prepareFunc: func() (string, error) {
				repoPath := "/tmp/repo_name"
				err := os.MkdirAll(repoPath, 0777)
				if err != nil {
					return "", err
				}
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return "", err
				}
				w, err := repo.Worktree()
				if err != nil {
					return "", err
				}
				//Apply some changes to default branch
				_, err = os.Create(repoPath + "/Readme.md")
				if err != nil {
					return "", err
				}
				_, err = os.Create(repoPath + "/file1")
				if err != nil {
					return "", err
				}
				_, err = w.Add(".")
				if err != nil {
					return "", err
				}
				_, err = w.Commit("initial commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				if err != nil {
					return "", err
				}

				//Create new branch to apply changes
				w, err = repo.Worktree()
				if err != nil {
					return "", err
				}
				err = w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("test-branch"),
					Create: true,
				})
				if err != nil {
					return "", err
				}

				err = os.Remove(repoPath + "/file1")
				if err != nil {
					return "", err
				}
				_, err = os.Create(repoPath + "/file2")
				if err != nil {
					return "", err
				}
				_, err = os.Create(repoPath + "/file3")
				if err != nil {
					return "", err
				}
				err = os.WriteFile(repoPath+"/Readme.md", []byte("Hi!"), 0777)
				if err != nil {
					return "", err
				}

				_, err = w.Add(".")
				if err != nil {
					return "", err
				}
				commitHash, err := w.Commit("commit to branch\ncommit message", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				return commitHash.String(), err
			},
			cleanupFunc: func() {
				os.RemoveAll("/tmp/repo_name")
				os.RemoveAll("/tmp/cloned_repo_name")
			},
			wantErr: false,
		},
		{
			name: "CloneRepoCommit fails when commit not exists",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
			},
			args: args{
				url:        "file:///tmp/repo_name",
				commitHash: "",
				repoPath:   "/tmp/cloned_repo_name",
			},
			prepareFunc: func() (string, error) {
				repoPath := "/tmp/repo_name"
				err := os.MkdirAll(repoPath, 0777)
				if err != nil {
					return "", err
				}
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return "", err
				}
				w, err := repo.Worktree()
				if err != nil {
					return "", err
				}
				//Apply some changes to default branch
				_, err = os.Create(repoPath + "/Readme.md")
				if err != nil {
					return "", err
				}
				_, err = os.Create(repoPath + "/file1")
				if err != nil {
					return "", err
				}
				_, err = w.Add(".")
				if err != nil {
					return "", err
				}
				_, err = w.Commit("initial commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				if err != nil {
					return "", err
				}
				return "some_random_hash", err
			},
			cleanupFunc: func() {
				os.RemoveAll("/tmp/repo_name")
				os.RemoveAll("/tmp/cloned_repo_name")
			},
			wantErr: true,
		},
		{
			name: "CloneRepoCommit fails when folder is not a git repo",
			fields: fields{
				logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})),
			},
			args: args{
				url:        "file:///tmp/repo_name",
				commitHash: "",
				repoPath:   "/tmp/cloned_repo_name",
			},
			prepareFunc: func() (string, error) {
				repoPath := "/tmp/repo_name"
				err := os.MkdirAll(repoPath, 0777)
				if err != nil {
					return "", err
				}
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return "", err
				}
				w, err := repo.Worktree()
				if err != nil {
					return "", err
				}
				//Apply some changes to default branch
				_, err = os.Create(repoPath + "/Readme.md")
				if err != nil {
					return "", err
				}
				_, err = os.Create(repoPath + "/file1")
				if err != nil {
					return "", err
				}
				_, err = w.Add(".")
				if err != nil {
					return "", err
				}

				return "some_invalid_commit", err
			},
			cleanupFunc: func() {
				os.RemoveAll("/tmp/repo_name")
				os.RemoveAll("/tmp/cloned_repo_name")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Git{
				logger: tt.fields.logger,
			}
			defer tt.cleanupFunc()
			commitHash, err := tt.prepareFunc()
			if err != nil {
				t.Errorf("Could not prepare test, error: %v", err.Error())
				return
			}
			tt.args.commitHash = commitHash
			if err := g.CloneRepoCommit(tt.args.url, tt.args.commitHash, tt.args.repoPath); (err != nil) != tt.wantErr {
				t.Errorf("CloneRepoCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if _, err := os.Stat(tt.args.repoPath + "/file1"); err == nil {
					t.Errorf("CloneRepoCommit() expected not to find file ./file1 but found it")
				}
				if _, err := os.Stat(tt.args.repoPath + "/file2"); err != nil {
					t.Errorf("CloneRepoCommit() expected to find file ./file2 but didn't find it")
				}
				if _, err := os.Stat(tt.args.repoPath + "/file3"); err != nil {
					t.Errorf("CloneRepoCommit() expected to find file ./file3 but didn't find it")
				}
			}
		})
	}
}
