package service

import (
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
)

type Service struct {
	UserService IUserService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	return &Service{
		UserService: NewUserService(repository.UserRepository),
	}
}
