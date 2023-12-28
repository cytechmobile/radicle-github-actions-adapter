package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log/slog"
)

type Git struct {
	logger *slog.Logger
}

func NewGit(logger *slog.Logger) *Git {
	return &Git{
		logger: logger,
	}
}

func (g *Git) CloneRepoCommit(url, commitHash, repoPath string) error {
	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})
	return err
}
