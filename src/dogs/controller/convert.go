package controller

import (
	"net/http"

	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/service"
)

type errRes struct {
	statusCode  int
	messageId   string
	typeDetails string
}

var errors map[error]errRes = make(map[error]errRes)

func getErrRes(err error) errRes {
	errResOk, exists := errors[err]
	if !exists {
		return errRes{
			statusCode: http.StatusInternalServerError,
			messageId:  "server.internal_error",
		}
	}

	return errResOk
}

func init() {
	errors[service.ErrDogNotExists] = errRes{
		statusCode: http.StatusNotFound,
		messageId:  "dog.not_found",
	}
	errors[service.ErrOwnerDogHasOneDog] = errRes{
		statusCode: http.StatusConflict,
		messageId:  "dog.has_owner",
	}
}
