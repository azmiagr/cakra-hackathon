package repository

import "gorm.io/gorm"

type Repository struct {
	UserRepository          IUserRepository
	CreditAccountRepository ICreditAccountRepository
	CategoryRepository      ICategoryRepository
	AnalysisRepository      IAnalysisRepository
	OtpRepository           IOtpRepository
	RegistrationRepository  IRegistrationSessionRepository
	RoleRepository          IRoleRepository
	PasswordResetRepository IPasswordResetRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:          NewUserRepository(db),
		CreditAccountRepository: NewCreditAccountRepository(db),
		CategoryRepository:      NewCategoryRepository(db),
		AnalysisRepository:      NewAnalysisRepository(db),
		OtpRepository:           NewOtpRepository(db),
		RegistrationRepository:  NewRegistrationSessionRepository(db),
		RoleRepository:          NewRoleRepository(db),
		PasswordResetRepository: NewPasswordResetRepository(db),
	}
}
