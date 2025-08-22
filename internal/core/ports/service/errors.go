package service

type ServiceErrKind string

func (e ServiceErrKind) String() string {
	return string(e)
}

const (
	Unexpected              ServiceErrKind = "unexpected error"
	FailedToAddPvz          ServiceErrKind = "failed to add pvz"
	ActiveReceptionExists   ServiceErrKind = "opened reception already exists"
	PvzNotFound             ServiceErrKind = "pvz not found"
	NoActiveReception       ServiceErrKind = "no opened reception"
	NoProdOrActiveReception ServiceErrKind = "no product to delete or active reception"
	FailedToCloseReception  ServiceErrKind = "failed to close reception"

	EmailAlreadyExists ServiceErrKind = "email already exists"
)
