package model

import "github.com/google/uuid"

type CreateAnalysisSessionRequest struct {
	CurrentStock int `json:"current_stock" binding:"gte=0"`
	LeadTimeDays int `json:"lead_time_days" binding:"gte=1,lte=365"`
}

type AnalysisUploadResponse struct {
	UploadID      uuid.UUID                       `json:"upload_id"`
	Status        string                          `json:"status"`
	SKUName       string                          `json:"sku_name,omitempty"`
	ValidRowCount int                             `json:"valid_row_count"`
	ErrorRowCount int                             `json:"error_row_count"`
	ValidRows     []AnalysisUploadPreviewRow      `json:"valid_rows"`
	Errors        []UploadValidationErrorResponse `json:"errors"`
}

type AnalysisUploadPreviewRow struct {
	RowNumber    int     `json:"row_number"`
	SaleDate     string  `json:"sale_date"`
	QuantitySold int     `json:"quantity_sold"`
	SKUName      string  `json:"sku_name"`
	UnitPrice    float64 `json:"unit_price"`
}

type UploadValidationErrorResponse struct {
	RowNumber int    `json:"row_number"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type AnalysisSessionResponse struct {
	SessionID        uuid.UUID               `json:"session_id"`
	Status           string                  `json:"status"`
	AvailableCredits int                     `json:"available_credits"`
	AIPayload        AIAnalysisPayload       `json:"ai_payload"`
	Recommendation   *RecommendationResponse `json:"recommendation,omitempty"`
	FailureCode      *string                 `json:"failure_code,omitempty"`
	FailureMessage   *string                 `json:"failure_message,omitempty"`
}

type RecommendationResponse struct {
	DemandCategory  string    `json:"demand_category"`
	ADIValue        *float64  `json:"adi_value,omitempty"`
	CVSquaredValue  *float64  `json:"cv_squared_value,omitempty"`
	ForecastP50     []float64 `json:"forecast_p50"`
	ForecastP90     []float64 `json:"forecast_p90"`
	ReorderPoint    int       `json:"reorder_point"`
	ReorderQuantity int       `json:"reorder_quantity"`
	RiskLabel       string    `json:"risk_label"`
	RiskReason      string    `json:"risk_reason"`
	ExplanationText string    `json:"explanation_text"`
}

type AIAnalysisPayload struct {
	AnalysisSessionID uuid.UUID           `json:"analysis_session_id"`
	SKUName           string              `json:"sku_name"`
	CurrentStock      int                 `json:"current_stock"`
	LeadTimeDays      int                 `json:"lead_time_days"`
	SalesHistory      []AISalesHistoryRow `json:"sales_history"`
}

type AISalesHistoryRow struct {
	SaleDate     string  `json:"sale_date"`
	QuantitySold int     `json:"quantity_sold"`
	UnitPrice    float64 `json:"unit_price"`
}

type AIResultRequest struct {
	Status          string    `json:"status" binding:"required,oneof=SUCCESS FAILED INSUFFICIENT_DATA"`
	DemandCategory  string    `json:"demand_category"`
	ADIValue        *float64  `json:"adi_value"`
	CVSquaredValue  *float64  `json:"cv_squared_value"`
	ForecastP50     []float64 `json:"forecast_p50"`
	ForecastP90     []float64 `json:"forecast_p90"`
	ReorderPoint    int       `json:"reorder_point"`
	ReorderQuantity int       `json:"reorder_quantity"`
	RiskLabel       string    `json:"risk_label"`
	RiskReason      string    `json:"risk_reason"`
	ExplanationText string    `json:"explanation_text"`
	ErrorCode       string    `json:"error_code"`
	ErrorMessage    string    `json:"error_message"`
}

type CreditAccountResponse struct {
	Balance          int `json:"balance"`
	ReservedCredits  int `json:"reserved_credits"`
	AvailableCredits int `json:"available_credits"`
}
