package radicle

import "context"

const CreatePatchCommentType = "revision.comment"

type CreatePatchComment struct {
	Type     string `json:"type"`
	Body     string `json:"body"`
	Revision string `json:"revision"`
}

// Patch should be implemented to support actions on Redicle patch
type Patch interface {
	Comment(ctx context.Context, repoID, patchID, revisionID string, message string) error
}
