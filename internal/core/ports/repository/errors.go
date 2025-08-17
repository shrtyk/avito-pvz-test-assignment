package repository

type RepoErrKind string

func (e RepoErrKind) String() string {
	return string(e)
}

const (
	KindFailed           RepoErrKind = "operation failed"
	KindNotFound         RepoErrKind = "entity not found"
	KindConflict         RepoErrKind = "entity conflicts with existing data"
	KindInvalidReference RepoErrKind = "invalid reference to another entity"
)
