package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	GetRoleByName(tx *gorm.DB, roleName string) (*entity.Role, error)
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetRoleByName(tx *gorm.DB, roleName string) (*entity.Role, error) {
	var role entity.Role

	err := tx.Where("role_name = ?", roleName).First(&role).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}
