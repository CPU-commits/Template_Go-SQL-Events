package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// https://www.rfc-editor.org/rfc/rfc9457.html
type ErrorProblemDetails struct {
	Title   string `json:"title"`
	Pointer string `json:"pointer"`
	Param   string `json:"param,omitempty"`
}

type ProblemDetails struct {
	Type   string                `json:"type,omitempty"`
	Title  string                `json:"title"`
	Detail string                `json:"detail,omitempty"`
	Errors []ErrorProblemDetails `json:"errors,omitempty"`
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "form.required"
	case "min":
		return "form.min"
	}
	return ""
}

func ValidatorErrorToErrorProblemDetails(
	err error,
	localizer *i18n.Localizer,
) []ErrorProblemDetails {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ErrorProblemDetails, len(ve))
		for i, fe := range ve {
			out[i] = ErrorProblemDetails{
				Pointer: fe.Field(),
				Title: localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: msgForTag(fe.Tag()),
				}),
				Param: fe.Param(),
			}
		}
		return out
	}
	return nil
}
