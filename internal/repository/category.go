package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICategoryRepository interface {
	GetOrCreate(tx *gorm.DB, category *entity.Category) (*entity.Category, error)
	ListByUserID(tx *gorm.DB, userID uuid.UUID) ([]entity.Category, error)
}

type CategoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) ICategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetOrCreate(tx *gorm.DB, category *entity.Category) (*entity.Category, error) {
	err := tx.
		Where("name = ? AND (user_id = ? OR user_id IS NULL)", category.Name, category.UserID).
		Order("user_id IS NULL ASC").
		First(category).Error
	if err == nil {
		return category, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	newCategory := *category
	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "name"}},
		DoNothing: true,
	}).Create(&newCategory).Error
	if err != nil {
		return nil, err
	}

	err = tx.
		Where("name = ? AND (user_id = ? OR user_id IS NULL)", category.Name, category.UserID).
		Order("user_id IS NULL ASC").
		First(category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) ListByUserID(tx *gorm.DB, userID uuid.UUID) ([]entity.Category, error) {
	items := make([]entity.Category, 0)
	err := tx.Where("user_id = ? OR user_id IS NULL", userID).
		Order("name ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
