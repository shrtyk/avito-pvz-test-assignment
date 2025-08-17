package repository

type RepoErrKind string

func (e RepoErrKind) String() string {
	return string(e)
}

const (
	KindUnexpected            RepoErrKind = "unexpected error"
	KindFailedInsertPvz       RepoErrKind = "failed to insert pvz"
	KindPvzNotFound           RepoErrKind = "pvz not found"
	KindActiveReceptionExists RepoErrKind = "in_progress reception already exists"
	KindNoActiveReception     RepoErrKind = "no in_progress reception"
)
