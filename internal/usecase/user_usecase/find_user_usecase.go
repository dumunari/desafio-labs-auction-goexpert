package user_usecase

import (
	"context"
	"fullcycle-auction_go/internal/entity/user_entity"
	"fullcycle-auction_go/internal/internal_error"
)

func NewUserUseCase(userRepository user_entity.UserRepositoryInterface) UserUseCaseInterface {
	return &UserUseCase{
		userRepository,
	}
}

type UserUseCase struct {
	UserRepository user_entity.UserRepositoryInterface
}

type UserOutputDTO struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UserInputDTO struct {
	Name string `json:"name" binding:"required,min=2"`
}

type UserUseCaseInterface interface {
	CreateUser(
		ctx context.Context,
		userInput UserInputDTO) (*UserOutputDTO, *internal_error.InternalError)

	FindUserById(
		ctx context.Context,
		id string) (*UserOutputDTO, *internal_error.InternalError)
}

func (u *UserUseCase) CreateUser(
	ctx context.Context,
	userInput UserInputDTO) (*UserOutputDTO, *internal_error.InternalError) {
	userEntity, err := user_entity.CreateUser(userInput.Name)
	if err != nil {
		return nil, err
	}

	if err := u.UserRepository.CreateUser(ctx, userEntity); err != nil {
		return nil, err
	}

	return &UserOutputDTO{
		Id:   userEntity.Id,
		Name: userEntity.Name,
	}, nil
}

func (u *UserUseCase) FindUserById(
	ctx context.Context, id string) (*UserOutputDTO, *internal_error.InternalError) {
	userEntity, err := u.UserRepository.FindUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserOutputDTO{
		Id:   userEntity.Id,
		Name: userEntity.Name,
	}, nil
}
