package service

import (
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/dto"
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/model"
	"github.com/CPU-commits/Template_Go-EventDriven/src/dogs/repository"
)

type dogService struct {
	dogRepository repository.DogRepository
}

func (dogService *dogService) GetDogById(idDog int) (*model.Dog, error) {
	dog, err := dogService.dogRepository.FindOne(&repository.CriteriaDog{
		ID: idDog,
	})
	if err != nil {
		return nil, err
	}
	if dog == nil {
		return nil, ErrDogNotExists
	}
	return dog, nil
}

func (dogService *dogService) InsertDog(dogDto dto.DogDTO) error {
	dog := dogDto.ToModel()

	exists, err := dogService.dogRepository.Exists(&repository.CriteriaDog{
		OwnerName: dog.OwnerName,
	})
	if err != nil {
		return err
	}
	if exists {
		return ErrOwnerDogHasOneDog
	}

	_, err = dogService.dogRepository.Insert(dog)
	return err
}

func NewDogService(
	dogRepository repository.DogRepository,
) *dogService {
	return &dogService{
		dogRepository: dogRepository,
	}
}
