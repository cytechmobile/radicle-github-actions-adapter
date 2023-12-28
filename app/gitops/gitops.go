package gitops

type Commit struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    Author   `json:"author"`
	Added     []string `json:"added"`
	Modified  []string `json:"modified"`
	Removed   []string `json:"removed"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitOps interface {
	CloneRepoCommit(url, commitHash, repoPath string) error
}
