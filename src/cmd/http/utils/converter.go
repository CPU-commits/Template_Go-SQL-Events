package utils

import (
	"net/http"

	"github.com/CPU-commits/Template_Go-EventDriven/src/auth/dto"
	authService "github.com/CPU-commits/Template_Go-EventDriven/src/auth/service"
	"github.com/CPU-commits/Template_Go-EventDriven/src/utils"
)

type errRes struct {
	StatusCode  int
	MessageId   string
	TypeDetails string
}

var errorsService map[error]errRes = make(map[error]errRes)

func GetErrRes(err error) errRes {
	errResOk, exists := errorsService[err]
	if !exists {
		return errRes{
			StatusCode: http.StatusInternalServerError,
			MessageId:  "server.internal_error",
		}
	}

	return errResOk
}

func init() {
	errorsService[authService.ErrInvalidCredentials] = errRes{
		StatusCode: http.StatusForbidden,
		MessageId:  "auth.err_credentials",
	}
	errorsService[authService.ErrUserLoginNotFound] = errRes{
		StatusCode: http.StatusForbidden,
		MessageId:  "auth.err_credentials",
	}
	errorsService[utils.ErrRepositoryFailed] = errRes{
		StatusCode: http.StatusServiceUnavailable,
		MessageId:  "server.db_error",
	}
	errorsService[authService.ErrSessionNotExists] = errRes{
		StatusCode: http.StatusConflict,
		MessageId:  "session.not_exists",
	}
	errorsService[authService.ErrUserNotFound] = errRes{
		StatusCode: http.StatusNotFound,
		MessageId:  "user.not_found",
	}
	errorsService[authService.ErrExistsEmail] = errRes{
		StatusCode: http.StatusConflict,
		MessageId:  "auth.exists_email",
	}
	errorsService[dto.ErrUnableToRegisterRole] = errRes{
		StatusCode: http.StatusBadRequest,
		MessageId:  "auth.unable_role",
	}
	errorsService[authService.ErrInvalidParams] = errRes{
		StatusCode: http.StatusBadRequest,
		MessageId:  "auth.invalid_params",
	}
}
