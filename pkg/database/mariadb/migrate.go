package mariadb

import (
	"github.com/azmiagr/cakra-hackathon/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.OtpCode{},
	)

	if err != nil {
		return err
	}

	return nil
}
