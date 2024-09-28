package controller

import (
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/repository"
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/service"
)

// Repositories
var (
	dogRepositoryMemory = repository.NewDogRepositoryMemory()
)

// Services
var (
	dogService = service.NewDogService(dogRepositoryMemory)
)
