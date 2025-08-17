package service

type ServiceErrKind string

func (e ServiceErrKind) String() string {
	return string(e)
}

const (
	KindUnexpected            ServiceErrKind = "unexpected error"
	KindFailedToAddPvz        ServiceErrKind = "failed to add pvz"
	KindActiveReceptionExists ServiceErrKind = "opened reception already exists"
	KindPvzNotFound           ServiceErrKind = "pvz not found"
	KindNoActiveReception     ServiceErrKind = "no opened reception"
)
