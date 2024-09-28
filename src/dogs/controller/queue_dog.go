package controller

import (
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/dto"
	"github.com/CPU-commits/Template_Go-EventDriven/src/package/bus"
)

func InsertDogQueue(c bus.Context) error {
	var dogDto dto.DogDTO
	if err := c.BindData(&dogDto); err != nil {
		return c.Kill(err.Error())
	}

	return dogService.InsertDog(dogDto)
}
