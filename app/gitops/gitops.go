package gitops

type GitOps interface {
	CloneRepoCommit(url, commitHash, repoPath string) error
}
