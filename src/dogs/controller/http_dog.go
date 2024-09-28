package controller

import (
	"net/http"
	"strconv"

	"github.com/CPU-commits/Template_Go-EventDriven/src/cmd/http/utils"
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/dto"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func GetDog(c *gin.Context) {
	idDog := c.Param("idDog")
	idIntDog, err := strconv.Atoi(idDog)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.ProblemDetails{
				Title: err.Error(),
			},
		)
		return
	}

	dog, err := dogService.GetDogById(idIntDog)
	if err != nil {
		localizer := utils.GetI18nLocalizer(c)
		errRes := getErrRes(err)
		c.AbortWithStatusJSON(
			errRes.statusCode,
			utils.ProblemDetails{
				Title: localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: errRes.messageId,
				}),
				Type: errRes.typeDetails,
			},
		)
		return
	}

	c.JSON(200, dog)
}

func InsertDog(c *gin.Context) {
	var dogDto *dto.DogDTO

	if err := c.BindJSON(&dogDto); err != nil {
		localizer := utils.GetI18nLocalizer(c)
		errors := utils.ValidatorErrorToErrorProblemDetails(err, localizer)

		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.ProblemDetails{
				Title: localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: "form.error",
				}),
				Errors: errors,
			},
		)
		return
	}
	if err := dogService.InsertDog(*dogDto); err != nil {
		localizer := utils.GetI18nLocalizer(c)
		errRes := getErrRes(err)
		c.AbortWithStatusJSON(
			errRes.statusCode,
			utils.ProblemDetails{
				Title: localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: errRes.messageId,
				}),
			},
		)
		return
	}

	c.JSON(201, utils.ProblemDetails{
		Title: "ok",
	})
}
