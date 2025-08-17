package repository

type RepoErrKind string

func (e RepoErrKind) String() string {
	return string(e)
}

const (
	Failed           RepoErrKind = "operation failed"
	NotFound         RepoErrKind = "entity not found"
	Conflict         RepoErrKind = "entity conflicts with existing data"
	InvalidReference RepoErrKind = "invalid reference to another entity"
)
