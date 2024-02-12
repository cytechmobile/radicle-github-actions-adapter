package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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

// CloneRepoCommit clones a repo from url to repoPath and checkouts to commitHash.
// It does not handle removing the created files.
func (g *Git) CloneRepoCommit(url, commitHash, repoPath string) error {
	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:               url,
		SingleBranch:      false,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		g.logger.Error("failed to clone repo from local storage", "url", url, "error", err.Error())
		return err
	}
	_, err = repo.Head()
	if err != nil {
		g.logger.Error("failed to get the head reference to repo", "error", err.Error())
		return err
	}

	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			"+refs/*:refs/*",
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		g.logger.Error("failed to fetch repo", "error", err.Error())
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		g.logger.Error("failed to get repo worktree", "error", err.Error())
		return err
	}

	// Checkout to the specific commit
	_, err = repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		g.logger.Error("failed to get commit hash", "error", err.Error())
		return err
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})
	if err != nil {
		g.logger.Error("failed to checkout to commit", "commit", commitHash, "error", err.Error())
		return err
	}
	_, err = repo.Head()
	if err != nil {
		g.logger.Error("failed to get commit worktree", "error", err.Error())
		return err
	}
	return nil
}
