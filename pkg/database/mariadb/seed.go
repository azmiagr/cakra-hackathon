package mariadb

import (
	"errors"

	"github.com/azmiagr/cakra-hackathon/entity"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var defaultCategoryNames = []string{
	"Makanan",
	"Minuman",
	"Kebutuhan Pokok",
	"Perawatan Tubuh",
	"Makanan Instan",
	"Kebersihan",
	"Bumbu Masak",
}

func Seed(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	return seedCategories(db)
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{RoleID: uuid.New(), RoleName: constants.RoleUser},
		{RoleID: uuid.New(), RoleName: constants.RoleAdmin},
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

func seedCategories(db *gorm.DB) error {
	for _, name := range defaultCategoryNames {
		var category entity.Category
		err := db.Where("user_id IS NULL AND name = ?", name).First(&category).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		category = entity.Category{CategoryID: uuid.New(), Name: name}
		if err := db.Create(&category).Error; err != nil {
			return err
		}
	}
	return nil
}
