package repository

type RepoErrKind string

func (e RepoErrKind) String() string {
	return string(e)
}

const (
	Unexpected       RepoErrKind = "unexpected error"
	FailedCreatePvz  RepoErrKind = "failed create pvz"
	NotFound         RepoErrKind = "entity not found"
	Conflict         RepoErrKind = "entity conflicts with existing data"
	InvalidReference RepoErrKind = "invalid reference to another entity"
)
