package radicle

import "context"

const CreatePatchCommentType = "revision.comment"
const EditPatchCommentType = "revision.comment.edit"

type CreatePatchComment struct {
	Type     string   `json:"type"`
	Body     string   `json:"body"`
	Revision string   `json:"revision"`
	Comment  *string  `json:"comment,omitempty"`
	Embeds   []string `json:"embeds"`
}

// Patch should be implemented to support actions on Redicle patch
type Patch interface {
	Comment(ctx context.Context, repoID, patchID, revisionID, message string, append bool) error
}
