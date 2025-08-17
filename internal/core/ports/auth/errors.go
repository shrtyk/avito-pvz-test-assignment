package auth

type AuthErrKind string

func (e AuthErrKind) String() string {
	return string(e)
}

const (
	JwtCreation      AuthErrKind = "failed jwt creation"
	InvalidJwt       AuthErrKind = "invalid jwt"
	ExpiredJwt       AuthErrKind = "jwt expired"
	NotAuthenticated AuthErrKind = "not authenticated"
	NotAuthorized    AuthErrKind = "not authorized"
	JwtClaimsFromCtx AuthErrKind = "failed to get JWT claims from context"
)
