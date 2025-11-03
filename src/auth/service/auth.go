package service

import (
	"strings"

	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/dto"
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/model"
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/repository/auth_repository"
	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/repository/user_repository"
	"github.com/CPU-commits/Template_Go-EventDriven/src/common/repository"
	"github.com/CPU-commits/Template_Go-EventDriven/src/package/bus"
	"github.com/CPU-commits/Template_Go-EventDriven/src/utils"
	"golang.org/x/crypto/bcrypt"
)

var authService *AuthService

type AuthService struct {
	authRepository auth_repository.AuthRepository
	userRepository user_repository.UserRepository
	bus            bus.Bus
}

func (authService *AuthService) CheckIfEmailExists(email, username string) error {
	existsEmailOrUsername, err := authService.userRepository.Exists(&user_repository.Criteria{
		Or: []user_repository.Criteria{
			{
				Email: repository.CriteriaString{
					EQ: utils.String(email),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if existsEmailOrUsername {
		return ErrExistsEmail
	}

	return nil
}

func (authService *AuthService) Register(
	registerDto *dto.RegisterDto,
	manualRegister bool,
) error {
	user, err := registerDto.ToModel()
	if err != nil {
		return err
	}
	err = authService.CheckIfEmailExists(registerDto.Email, registerDto.Username)
	if err != nil {
		return err
	}
	var auth *model.Auth
	if registerDto.Auth != nil && registerDto.Password == "" {
		auth = registerDto.Auth
	} else if registerDto.Password != "" {
		auth = &model.Auth{
			Password:   registerDto.Password,
			Provider:   "local",
			ProviderId: registerDto.Email,
		}
	}

	_, err = authService.userRepository.InsertOne(user, auth)
	return err
}

func (authService *AuthService) RegisterOrLoginExternalAuth(
	externalAuthDto *dto.ExternalAuthDTO,
) (*model.User, int64, error) {
	user, err := authService.userRepository.FindOneByEmail(externalAuthDto.Email)
	if err != nil {
		return nil, 0, utils.ErrRepositoryFailed
	}
	if user == nil {
		username := strings.Split(externalAuthDto.Email, "@")[0]
		if err := authService.Register(&dto.RegisterDto{
			Name:     externalAuthDto.Name,
			Username: username,
			Email:    externalAuthDto.Email,
			Auth: &model.Auth{
				Provider:   externalAuthDto.Provider,
				ProviderId: externalAuthDto.Email,
			},
		}, false); err != nil {
			return nil, 0, err
		}

		user, err = authService.userRepository.FindOneByEmail(externalAuthDto.Email)
		if err != nil {
			return nil, 0, utils.ErrRepositoryFailed
		}
		if externalAuthDto.AvatarURL != "" {
			// External avatar do
		}
	}
	existsProvider, err := authService.authRepository.Exists(&auth_repository.Criteria{
		Provider: externalAuthDto.Provider,
		IDUser:   user.ID,
	})
	if err != nil {
		return nil, 0, err
	}
	if !existsProvider {
		idAuth, err := authService.authRepository.InsertOne(&model.Auth{
			Provider: externalAuthDto.Provider,
			Password: externalAuthDto.Email,
			IDUser:   user.ID,
		})
		if err != nil {
			return nil, 0, err
		}

		return user, idAuth, nil
	}

	auth, err := authService.authRepository.FindOneByUsernameAndProvider(user.Email, externalAuthDto.Provider)
	if err != nil {
		return nil, 0, utils.ErrRepositoryFailed
	}

	return user, auth.ID, nil
}

func (authService *AuthService) LoginLocal(authDto dto.AuthDto) (*model.User, int64, error) {
	auth, err := authService.authRepository.FindOneByUsernameAndProvider(authDto.Username, "local")
	if err != nil {
		return nil, 0, utils.ErrRepositoryFailed
	}
	if auth == nil {
		return nil, 0, ErrUserLoginNotFound
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(auth.Password),
		[]byte(authDto.Password),
	); err != nil {
		return nil, 0, ErrInvalidCredentials
	}

	// User
	user, err := authService.userRepository.FindOneByEmail(authDto.Username)
	if err != nil {
		return nil, 0, err
	}
	if user == nil {
		return nil, 0, ErrUserLoginNotFound
	}

	return user, auth.ID, nil
}

func NewAuthService(
	authRepository auth_repository.AuthRepository,
	userRepository user_repository.UserRepository,
	bus bus.Bus,
) *AuthService {
	if authService == nil {
		authService = &AuthService{
			authRepository: authRepository,
			userRepository: userRepository,
			bus:            bus,
		}
	}
	return authService
}
