package utils

import (
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators() {
	if _, ok := binding.Validator.Engine().(*validator.Validate); ok {
		fmt.Printf("ok: %v\n", ok)
	}
}
