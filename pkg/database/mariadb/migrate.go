package mariadb

import (
	"github.com/azmiagr/cakra-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.CreditAccount{},
		&entity.CreditTransaction{},
		&entity.AnalysisUpload{},
		&entity.AnalysisUploadRow{},
		&entity.UploadValidationError{},
		&entity.Category{},
		&entity.SKU{},
		&entity.AnalysisSession{},
		&entity.SalesHistory{},
		&entity.RecommendationResult{},
		&entity.OtpCode{},
		&entity.RegistrationSession{},
		&entity.PasswordReset{},
	)

	if err != nil {
		return err
	}
	return nil
}
