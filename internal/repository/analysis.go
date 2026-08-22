package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAnalysisRepository interface {
	CreateUpload(tx *gorm.DB, upload *entity.AnalysisUpload) error
	CreateUploadRows(tx *gorm.DB, rows []entity.AnalysisUploadRow) error
	CreateValidationErrors(tx *gorm.DB, rows []entity.UploadValidationError) error
	GetUploadOwned(tx *gorm.DB, uploadID, userID uuid.UUID) (*entity.AnalysisUpload, error)
	ListUploadRows(tx *gorm.DB, uploadID uuid.UUID) ([]entity.AnalysisUploadRow, error)
	ListValidationErrors(tx *gorm.DB, uploadID uuid.UUID) ([]entity.UploadValidationError, error)
	GetOrCreateSKU(tx *gorm.DB, sku *entity.SKU) (*entity.SKU, error)
	CreateSession(tx *gorm.DB, session *entity.AnalysisSession) error
	CreateSalesHistories(tx *gorm.DB, rows []entity.SalesHistory) error
	GetSessionOwned(tx *gorm.DB, sessionID, userID uuid.UUID) (*entity.AnalysisSession, *entity.SKU, *entity.RecommendationResult, error)
	GetSessionForUpdate(tx *gorm.DB, sessionID uuid.UUID) (*entity.AnalysisSession, error)
	ListSalesHistories(tx *gorm.DB, sessionID uuid.UUID) ([]entity.SalesHistory, error)
	CreateRecommendationResult(tx *gorm.DB, result *entity.RecommendationResult) error
	UpdateSessionResult(tx *gorm.DB, sessionID uuid.UUID, status string, demandCategory *string, adiValue, cvSquaredValue *float64, failureCode, failureMessage *string) error
}

type AnalysisRepository struct{ db *gorm.DB }

func NewAnalysisRepository(db *gorm.DB) IAnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) CreateUpload(tx *gorm.DB, upload *entity.AnalysisUpload) error {
	err := tx.Create(upload).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AnalysisRepository) CreateUploadRows(tx *gorm.DB, rows []entity.AnalysisUploadRow) error {
	if len(rows) == 0 {
		return nil
	}

	err := tx.Create(&rows).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AnalysisRepository) CreateValidationErrors(tx *gorm.DB, rows []entity.UploadValidationError) error {
	if len(rows) == 0 {
		return nil
	}

	err := tx.Create(&rows).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AnalysisRepository) GetUploadOwned(tx *gorm.DB, uploadID, userID uuid.UUID) (*entity.AnalysisUpload, error) {
	var row entity.AnalysisUpload

	err := tx.Where("analysis_upload_id = ? AND user_id = ?", uploadID, userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AnalysisRepository) ListUploadRows(tx *gorm.DB, uploadID uuid.UUID) ([]entity.AnalysisUploadRow, error) {
	var rows []entity.AnalysisUploadRow

	err := tx.Where("analysis_upload_id = ?", uploadID).Order("row_number ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AnalysisRepository) ListValidationErrors(tx *gorm.DB, uploadID uuid.UUID) ([]entity.UploadValidationError, error) {
	var rows []entity.UploadValidationError

	err := tx.Where("analysis_upload_id = ?", uploadID).Order("row_number ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AnalysisRepository) GetOrCreateSKU(tx *gorm.DB, sku *entity.SKU) (*entity.SKU, error) {
	var existing entity.SKU

	err := tx.Where("user_id = ? AND name = ?", sku.UserID, sku.Name).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	err = tx.Create(sku).Error
	if err != nil {
		return nil, err
	}

	return sku, nil
}
func (r *AnalysisRepository) CreateSession(tx *gorm.DB, session *entity.AnalysisSession) error {
	err := tx.Create(session).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AnalysisRepository) CreateSalesHistories(tx *gorm.DB, rows []entity.SalesHistory) error {
	if len(rows) == 0 {
		return nil
	}

	err := tx.Create(&rows).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AnalysisRepository) GetSessionOwned(tx *gorm.DB, sessionID, userID uuid.UUID) (*entity.AnalysisSession, *entity.SKU, *entity.RecommendationResult, error) {
	var session entity.AnalysisSession

	err := tx.Where("analysis_session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return nil, nil, nil, err
	}

	var sku entity.SKU
	err = tx.Where("sku_id = ?", session.SKUID).First(&sku).Error
	if err != nil {
		return nil, nil, nil, err
	}

	var result entity.RecommendationResult
	err = tx.Where("analysis_session_id = ?", sessionID).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return &session, &sku, nil, nil
	}
	if err != nil {
		return nil, nil, nil, err
	}

	return &session, &sku, &result, nil
}
func (r *AnalysisRepository) GetSessionForUpdate(tx *gorm.DB, sessionID uuid.UUID) (*entity.AnalysisSession, error) {
	var session entity.AnalysisSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("analysis_session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}
func (r *AnalysisRepository) ListSalesHistories(tx *gorm.DB, sessionID uuid.UUID) ([]entity.SalesHistory, error) {
	var rows []entity.SalesHistory
	err := tx.Where("analysis_session_id = ?", sessionID).Order("sale_date ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}
func (r *AnalysisRepository) CreateRecommendationResult(tx *gorm.DB, result *entity.RecommendationResult) error {
	err := tx.Create(result).Error
	if err != nil {
		return err
	}
	return nil
}
func (r *AnalysisRepository) UpdateSessionResult(tx *gorm.DB, sessionID uuid.UUID, status string, demandCategory *string, adiValue, cvSquaredValue *float64, failureCode, failureMessage *string) error {
	err := tx.Model(&entity.AnalysisSession{}).Where("analysis_session_id = ?", sessionID).Updates(map[string]any{"status": status, "demand_category": demandCategory, "adi_value": adiValue, "cv_squared_value": cvSquaredValue, "failure_code": failureCode, "failure_message": failureMessage}).Error
	if err != nil {
		return err
	}
	return nil
}
