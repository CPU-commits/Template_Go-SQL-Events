package user_repository

import (
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"
	"github.com/CPU-commits/Template_Go-EventDriven/src/common/repository"
)

type Criteria struct {
	ID     int64
	ID_NIN []int64
	Name   repository.CriteriaString
	Email  repository.CriteriaString
	Or     []Criteria
	Roles  []model.Role
}

type UserRepository interface {
	FindOneByEmail(email string) (*model.User, error)
	FindOneByID(id int64) (*model.User, error)
	Exists(criteria *Criteria) (bool, error)
	InsertOne(user *model.User, auth *model.Auth) (*model.User, error)
}
