package pwdservice

//go:generate  mockery
type PasswordService interface {
	Hash(plainPwd string) ([]byte, error)
	Compare(hashedPwd []byte, plainPwd string) (bool, error)
}
