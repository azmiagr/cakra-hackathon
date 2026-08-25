package service

import (
	"github.com/azmiagr/cakra-hackathon/internal/repository"
	"github.com/azmiagr/cakra-hackathon/pkg/ai"
	"github.com/azmiagr/cakra-hackathon/pkg/bcrypt"
	"github.com/azmiagr/cakra-hackathon/pkg/config"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
	"github.com/azmiagr/cakra-hackathon/pkg/mail"
	"github.com/azmiagr/cakra-hackathon/pkg/supabase"
)

type Service struct {
	UserService     IUserService
	AuthService     IAuthService
	AnalysisService IAnalysisService
	CreditService   ICreditService
}

func NewService(repository *repository.Repository, bcryptAuth bcrypt.Interface, jwtAuth jwt.Interface, mailer mail.Interface, registrationConf config.RegistrationConfig, storage supabase.Interface, predictor ai.Predictor) *Service {
	return &Service{
		UserService:     NewUserService(repository.UserRepository),
		AuthService:     NewAuthService(repository.UserRepository, repository.CreditAccountRepository, repository.OtpRepository, repository.RegistrationRepository, repository.PasswordResetRepository, repository.RoleRepository, bcryptAuth, jwtAuth, mailer, registrationConf),
		AnalysisService: NewAnalysisService(repository.AnalysisRepository, repository.CategoryRepository, repository.CreditAccountRepository, storage, predictor),
		CreditService:   NewCreditService(repository.CreditAccountRepository),
	}
}
