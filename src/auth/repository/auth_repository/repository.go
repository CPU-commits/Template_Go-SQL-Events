package auth_repository

import "github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"

type Criteria struct {
	Provider string
	IDUser   int64
}

type DataUpdate struct {
	Password *string
}

type AuthRepository interface {
	FindOneByUsernameAndProvider(username string, provider string) (*model.Auth, error)
	FindOneByUserId(userId int64) (*model.Auth, error)
	FindOneByUserIdAndProvider(userId int64, provider string) (*model.Auth, error)
	UpdatePassword(userId int64, data DataUpdate) error
	Exists(criteria *Criteria) (bool, error)
	InsertOne(auth *model.Auth) (int64, error)
}
