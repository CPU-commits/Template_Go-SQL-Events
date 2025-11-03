package token

import (
	"time"

	userModel "github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"
)

type TokenGenerator interface {
	NewSessionToken(expiredAt time.Time, idUser int64) (string, error)
	NewAccessToken(expiredAt time.Time, user userModel.User) (string, error)
	NewFirstTimeToken(IDUser int64) (string, error)
	NewRecoveryToken(expiredAt time.Time, user userModel.User) (string, error)
}
