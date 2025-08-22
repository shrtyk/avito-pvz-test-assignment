package pwdservice

type PwdErrKind string

func (e PwdErrKind) String() string {
	return string(e)
}

const (
	Unexpected    PwdErrKind = "unexpected error"
	WrongPassword PwdErrKind = "wrong password"
)
