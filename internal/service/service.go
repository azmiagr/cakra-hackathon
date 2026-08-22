package service

import (
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/config"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
	"github.com/azmiagr/cakra-hackathon/pkg/mail"
	"gorm.io/gorm"
)

type Service struct {
	UserService IUserService
	AuthService IAuthService
}

func NewService(
	db *gorm.DB,
	repository *repository.Repository,
	bcryptAuth bcrypt.Interface,
	jwtAuth jwt.Interface,
	mailer mail.Interface,
	registrationConf config.RegistrationConfig,
) *Service {
	return &Service{
		UserService: NewUserService(db, repository.UserRepository),
		AuthService: NewAuthService(
			db,
			repository.UserRepository,
			repository.OtpRepository,
			repository.RegistrationRepository,
			repository.PasswordResetRepository,
			repository.RoleRepository,
			bcryptAuth,
			jwtAuth,
			mailer,
			registrationConf,
		),
	}
}
