package mariadb

import (
	"errors"

	"github.com/azmiagr/cakra-hackathon/entity"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	err := seedRoles(db)
	if err != nil {
		return err
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{
			RoleID:   uuid.New(),
			RoleName: constants.RoleUser,
		},
		{
			RoleID:   uuid.New(),
			RoleName: constants.RoleAdmin,
		},
	}

	for _, role := range roles {
		var existing entity.Role
		err := db.Where("role_name = ?", role.RoleName).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	return nil
}
