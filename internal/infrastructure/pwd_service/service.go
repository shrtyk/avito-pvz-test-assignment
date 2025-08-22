package pwdservice

import (
	"errors"

	pwdservice "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/pwd_service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
	"github.com/tailscale/golang-x-crypto/bcrypt"
)

type pwdService struct{}

func NewPasswordService() *pwdService {
	return &pwdService{}
}

func (s *pwdService) Hash(plainPwd string) ([]byte, error) {
	op := "pwd_service.Hash"

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPwd), 12)
	if err != nil {
		return nil, xerr.WrapErr(op, pwdservice.Unexpected, err)
	}

	return hash, nil
}

func (s *pwdService) Compare(hashedPwd []byte, plainPwd string) (bool, error) {
	op := "pwd_service.Compare"

	err := bcrypt.CompareHashAndPassword(hashedPwd, []byte(plainPwd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, xerr.WrapErr(op, pwdservice.WrongPassword, err)
		}
	}
	return true, nil
}
