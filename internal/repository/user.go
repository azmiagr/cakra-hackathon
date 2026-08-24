package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/azmiagr/cakra-hackathon/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error)
	CreateUser(tx *gorm.DB, user *entity.User) error
	UpdateUser(tx *gorm.DB, user *entity.User) error
	ActivateUser(tx *gorm.DB, userID uuid.UUID, passwordHash string) error
	IncrementSessionVersion(tx *gorm.DB, userID uuid.UUID) error
	UpdatePasswordAndIncrementSessionVersion(tx *gorm.DB, userID uuid.UUID, passwordHash string) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(tx *gorm.DB, param model.GetUserParam) (*entity.User, error) {
	var user entity.User
	query := tx.Where(&param)
	if param.Email != "" {
		query = query.Where("email = ?", param.Email)
	}
	if param.UserID != uuid.Nil {
		query = query.Where("user_id = ?", param.UserID)
	}

	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(tx *gorm.DB, user *entity.User) error {
	err := tx.Save(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) ActivateUser(tx *gorm.DB, userID uuid.UUID, passwordHash string) error {
	err := tx.Model(&entity.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"password": passwordHash,
			"status":   "active",
		}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) IncrementSessionVersion(tx *gorm.DB, userID uuid.UUID) error {
	err := tx.Model(&entity.User{}).
		Where("user_id = ?", userID).
		Update("session_version", gorm.Expr("session_version + ?", 1)).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdatePasswordAndIncrementSessionVersion(tx *gorm.DB, userID uuid.UUID, passwordHash string) error {
	err := tx.Model(&entity.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"password":        passwordHash,
			"session_version": gorm.Expr("session_version + ?", 1),
		}).Error
	if err != nil {
		return err
	}
	return nil
}
