package pwdservice

import (
	"errors"
	"fmt"

	"github.com/tailscale/golang-x-crypto/bcrypt"
)

type pwdService struct{}

func NewPasswordService() *pwdService {
	return &pwdService{}
}

func (s *pwdService) Hash(plainPwd string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPwd), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash out of password: %w", err)
	}

	return hash, nil
}

func (s *pwdService) Compare(hashedPwd []byte, plainPwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashedPwd, []byte(plainPwd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, fmt.Errorf("failed to compare passwords: %w", err)
		}
	}
	return true, nil
}
