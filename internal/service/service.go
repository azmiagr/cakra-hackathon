package service

import (
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
)

type Service struct {
	UserService IUserService
	OtpService  IOtpService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface) *Service {
	userService := NewUserService(repository.UserRepository)
	otpService := NewOtpService(repository.OtpRepository, repository.UserRepository)

	return &Service{
		UserService: userService,
		OtpService:  otpService,
	}
}
