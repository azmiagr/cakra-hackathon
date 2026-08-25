package repository

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/azmiagr/cakra-hackathon/model"
	constants "github.com/azmiagr/cakra-hackathon/pkg/constant"
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
	SetSKUCategory(tx *gorm.DB, skuID, categoryID uuid.UUID) error
	CreateSession(tx *gorm.DB, session *entity.AnalysisSession) error
	CreateSalesHistories(tx *gorm.DB, rows []entity.SalesHistory) error
	GetSessionOwned(tx *gorm.DB, sessionID, userID uuid.UUID) (*entity.AnalysisSession, *entity.SKU, *entity.Category, *entity.RecommendationResult, error)
	GetSessionForUpdate(tx *gorm.DB, sessionID uuid.UUID) (*entity.AnalysisSession, error)
	ListSalesHistories(tx *gorm.DB, sessionID uuid.UUID) ([]entity.SalesHistory, error)
	CreateRecommendationResult(tx *gorm.DB, result *entity.RecommendationResult) error
	UpdateSessionResult(tx *gorm.DB, sessionID uuid.UUID, status string, demandCategory *string, adiValue, cvSquaredValue *float64, failureCode, failureMessage *string) error
	ListHistory(tx *gorm.DB, userID uuid.UUID, query model.AnalysisHistoryQuery) ([]model.AnalysisHistoryItem, int, error)
	GetHistorySummary(tx *gorm.DB, userID uuid.UUID) (*model.AnalysisHistorySummary, error)
	CountCompletedSKUs(tx *gorm.DB, userID uuid.UUID) (int, error)
	ListDashboardAlerts(tx *gorm.DB, userID uuid.UUID, limit int) ([]model.DashboardAlert, error)
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

	err := tx.
		Where("analysis_upload_id = ?", uploadID).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "row_number"}}).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AnalysisRepository) ListValidationErrors(tx *gorm.DB, uploadID uuid.UUID) ([]entity.UploadValidationError, error) {
	var rows []entity.UploadValidationError

	err := tx.
		Where("analysis_upload_id = ?", uploadID).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "row_number"}}).
		Find(&rows).Error
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

func (r *AnalysisRepository) SetSKUCategory(tx *gorm.DB, skuID, categoryID uuid.UUID) error {
	err := tx.Model(&entity.SKU{}).
		Where("sku_id = ?", skuID).
		Update("category_id", categoryID).
		Error
	if err != nil {
		return err
	}
	return nil
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

func (r *AnalysisRepository) GetSessionOwned(tx *gorm.DB, sessionID, userID uuid.UUID) (*entity.AnalysisSession, *entity.SKU, *entity.Category, *entity.RecommendationResult, error) {
	var session entity.AnalysisSession

	err := tx.Where("analysis_session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var sku entity.SKU
	err = tx.Where("sku_id = ?", session.SKUID).First(&sku).Error
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var category *entity.Category
	if sku.CategoryID != nil {
		category = &entity.Category{}
		err = tx.Where("category_id = ?", *sku.CategoryID).First(category).Error
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	var result entity.RecommendationResult
	err = tx.Where("analysis_session_id = ?", sessionID).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return &session, &sku, category, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return &session, &sku, category, &result, nil
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

func (r *AnalysisRepository) ListHistory(tx *gorm.DB, userID uuid.UUID, query model.AnalysisHistoryQuery) ([]model.AnalysisHistoryItem, int, error) {
	base := tx.
		Table("analysis_sessions").
		Joins("JOIN skus ON skus.sku_id = analysis_sessions.sku_id").
		Joins("LEFT JOIN categories ON categories.category_id = skus.category_id").
		Joins("JOIN recommendation_results ON recommendation_results.analysis_session_id = analysis_sessions.analysis_session_id").
		Where("analysis_sessions.user_id = ? AND analysis_sessions.status = ?", userID, constants.AnalysisSessionCompleted)

	if query.Search != "" {
		base = base.Where("skus.name LIKE ?", "%"+query.Search+"%")
	}
	if query.RiskLabel != "" {
		base = base.Where("recommendation_results.risk_label = ?", query.RiskLabel)
	}

	var total int64
	err := base.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	order := "recommendation_results.created_at DESC"
	if query.Sort == "oldest" {
		order = "recommendation_results.created_at ASC"
	}

	items := make([]model.AnalysisHistoryItem, 0)
	err = base.
		Select("analysis_sessions.analysis_session_id AS session_id, skus.sku_id, skus.name AS sku_name, categories.name AS category, analysis_sessions.status AS session_status, recommendation_results.risk_label, recommendation_results.reorder_point, recommendation_results.reorder_quantity, DATE_FORMAT(recommendation_results.created_at, '%Y-%m-%d') AS analysis_date").
		Order(order).
		Limit(query.Limit).
		Offset((query.Page - 1) * query.Limit).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

func (r *AnalysisRepository) GetHistorySummary(tx *gorm.DB, userID uuid.UUID) (*model.AnalysisHistorySummary, error) {
	summary := &model.AnalysisHistorySummary{}

	var totalAnalysis int64
	err := tx.
		Table("analysis_sessions").
		Where("user_id = ? AND status = ?", userID, constants.AnalysisSessionCompleted).
		Count(&totalAnalysis).Error
	if err != nil {
		return nil, err
	}
	summary.TotalAnalysis = int(totalAnalysis)

	var atRiskSKUCount int64
	err = tx.
		Table("analysis_sessions").
		Joins("JOIN recommendation_results ON recommendation_results.analysis_session_id = analysis_sessions.analysis_session_id").
		Where("analysis_sessions.user_id = ? AND analysis_sessions.status = ? AND recommendation_results.risk_label <> ?", userID, constants.AnalysisSessionCompleted, "NORMAL").
		Distinct("analysis_sessions.sku_id").
		Count(&atRiskSKUCount).Error
	if err != nil {
		return nil, err
	}
	summary.AtRiskSKUCount = int(atRiskSKUCount)

	return summary, nil
}

func (r *AnalysisRepository) CountCompletedSKUs(tx *gorm.DB, userID uuid.UUID) (int, error) {
	var count int64
	err := tx.
		Model(&entity.AnalysisSession{}).
		Where("user_id = ? AND status = ?", userID, constants.AnalysisSessionCompleted).
		Distinct("sku_id").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *AnalysisRepository) ListDashboardAlerts(tx *gorm.DB, userID uuid.UUID, limit int) ([]model.DashboardAlert, error) {
	alerts := make([]model.DashboardAlert, 0)
	err := tx.
		Table("analysis_sessions").
		Joins("JOIN skus ON skus.sku_id = analysis_sessions.sku_id").
		Joins("JOIN recommendation_results ON recommendation_results.analysis_session_id = analysis_sessions.analysis_session_id").
		Where("analysis_sessions.user_id = ? AND analysis_sessions.status = ? AND recommendation_results.risk_label <> ?", userID, constants.AnalysisSessionCompleted, "NORMAL").
		Select("analysis_sessions.analysis_session_id AS session_id, skus.sku_id, skus.name AS sku_name, recommendation_results.risk_label, recommendation_results.risk_reason, DATE_FORMAT(recommendation_results.created_at, '%Y-%m-%d') AS analysis_date").
		Order("recommendation_results.created_at DESC").
		Limit(limit).
		Scan(&alerts).Error
	if err != nil {
		return nil, err
	}
	return alerts, nil
}
