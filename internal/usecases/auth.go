package usecases

import (
	"github.com/abayken/yandex-practicum-diploma/internal/custom_errors"
	"github.com/abayken/yandex-practicum-diploma/internal/repositories"
)

type AuthUseCase struct {
	Repository repositories.AuthRepository
}

func (usecase AuthUseCase) Register(login, password string) error {
	exists, err := usecase.Repository.Exists(login)

	if err != nil {
		return err
	}

	if exists {
		return &custom_errors.AlreadyExistsUserError{}
	} else {
		return nil
	}
}
