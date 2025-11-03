package service

import (
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/repository/user_repository"
)

var userService *UserService

type UserService struct {
	userRepository user_repository.UserRepository
}

func (userService *UserService) GetUserById(idUser int64) (*model.User, error) {
	user, err := userService.userRepository.FindOneByID(idUser)
	if err != nil {
		return nil, nil
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func NewUserService(
	userRepository user_repository.UserRepository,
) *UserService {
	if userService == nil {
		userService = &UserService{
			userRepository: userRepository,
		}
	}
	return userService
}
