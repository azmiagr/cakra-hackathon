package model

import "github.com/google/uuid"

type CreateAnalysisSessionRequest struct {
	CategoryName string `json:"category_name" binding:"required,max=100"`
	CurrentStock int    `json:"current_stock" binding:"gte=0"`
	LeadTimeDays int    `json:"lead_time_days" binding:"gte=1,lte=365"`
}

type CategoryResponse struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
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
	AIPayload        *AIAnalysisPayload      `json:"ai_payload,omitempty"`
	Recommendation   *RecommendationResponse `json:"recommendation,omitempty"`
	Result           *AnalysisResultResponse `json:"result,omitempty"`
	FailureCode      *string                 `json:"failure_code,omitempty"`
	FailureMessage   *string                 `json:"failure_message,omitempty"`
}

type AnalysisHistoryQuery struct {
	Search    string
	RiskLabel string
	Page      int
	Limit     int
	Sort      string
}

type AnalysisHistoryResponse struct {
	Summary    AnalysisHistorySummary `json:"summary"`
	Items      []AnalysisHistoryItem  `json:"items"`
	Pagination PaginationResponse     `json:"pagination"`
}

type AnalysisHistorySummary struct {
	TotalAnalysis   int      `json:"total_analysis"`
	AverageAccuracy *float64 `json:"average_accuracy"`
	AccuracyReady   bool     `json:"accuracy_ready"`
	AtRiskSKUCount  int      `json:"at_risk_sku_count"`
}

type AnalysisHistoryItem struct {
	SessionID       uuid.UUID `json:"session_id"`
	SKUID           uuid.UUID `json:"sku_id"`
	SKUName         string    `json:"sku_name"`
	Category        *string   `json:"category"`
	SessionStatus   string    `json:"session_status"`
	RiskLabel       string    `json:"risk_label"`
	ReorderPoint    int       `json:"reorder_point"`
	ReorderQuantity int       `json:"reorder_quantity"`
	AnalysisDate    string    `json:"analysis_date"`
}

type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type DashboardResponse struct {
	UserName            string                `json:"user_name"`
	TotalAnalyzedSKUs   int                   `json:"total_analyzed_skus"`
	AvailableCredits    int                   `json:"available_credits"`
	StockoutRiskCount   int                   `json:"stockout_risk_count"`
	AverageAccuracy     *float64              `json:"average_accuracy"`
	AccuracyReady       bool                  `json:"accuracy_ready"`
	RecentAnalyses      []AnalysisHistoryItem `json:"recent_analyses"`
	UrgentSKUs          []DashboardAlert      `json:"urgent_skus"`
	CreditAccount       CreditAccountResponse `json:"credit_account"`
	CreditUsedThisMonth int                   `json:"credit_used_this_month"`
}

type DashboardAlert struct {
	SessionID    uuid.UUID `json:"session_id"`
	SKUID        uuid.UUID `json:"sku_id"`
	SKUName      string    `json:"sku_name"`
	RiskLabel    string    `json:"risk_label"`
	RiskReason   string    `json:"risk_reason"`
	AnalysisDate string    `json:"analysis_date"`
}

type AnalysisResultResponse struct {
	SKU                AnalysisResultSKU     `json:"sku"`
	AnalysisDate       string                `json:"analysis_date"`
	HistoricalData     HistoricalDataSummary `json:"historical_data"`
	CurrentStock       int                   `json:"current_stock"`
	LeadTimeDays       int                   `json:"lead_time_days"`
	TargetServiceLevel float64               `json:"target_service_level"`
	DemandCategory     string                `json:"demand_category"`
	AverageDailyDemand float64               `json:"average_daily_demand"`
	Forecast           ForecastResponse      `json:"forecast"`
	ReorderPoint       int                   `json:"reorder_point"`
	ReorderQuantity    int                   `json:"reorder_quantity"`
	Risk               RiskResponse          `json:"risk"`
	ExplanationText    string                `json:"explanation_text"`
}

type AnalysisResultSKU struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type HistoricalDataSummary struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	PeriodDays int    `json:"period_days"`
	RowCount   int    `json:"row_count"`
}

type ForecastResponse struct {
	HorizonDays int             `json:"horizon_days"`
	Points      []ForecastPoint `json:"points"`
}

type ForecastPoint struct {
	Date string  `json:"date"`
	P50  float64 `json:"p50"`
	P90  float64 `json:"p90"`
}

type RiskResponse struct {
	Label  string `json:"label"`
	Reason string `json:"reason"`
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
