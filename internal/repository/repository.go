package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository          IUserRepository
	OtpRepository           IOtpRepository
	RegistrationRepository  IRegistrationSessionRepository
	RoleRepository          IRoleRepository
	PasswordResetRepository IPasswordResetRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:          NewUserRepository(db),
		OtpRepository:           NewOtpRepository(db),
		RegistrationRepository:  NewRegistrationSessionRepository(db),
		RoleRepository:          NewRoleRepository(db),
		PasswordResetRepository: NewPasswordResetRepository(db),
	}
}
