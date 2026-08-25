package ai

import "context"

type Predictor interface {
	Predict(ctx context.Context, input PredictionInput) (PredictionOutput, error)
}

type PredictionInput struct {
	SKUName      string              `json:"sku_name"`
	CategoryName string              `json:"category_name"`
	CurrentStock int                 `json:"current_stock"`
	LeadTimeDays int                 `json:"lead_time_days"`
	SessionID    string              `json:"session_id"`
	SalesHistory []SalesHistoryPoint `json:"sales_history"`
}

type SalesHistoryPoint struct {
	Date         string `json:"date"`
	QuantitySold int    `json:"quantity_sold"`
}

type PredictionOutput struct {
	Status              string
	DemandCategory      string
	ADIValue            *float64
	CVSquaredValue      *float64
	ForecastHorizonDays int
	ForecastP50         []float64
	ForecastP90         []float64
	ReorderPoint        int
	ReorderQuantity     int
	RiskLabel           string
	RiskReason          string
	ExplanationText     string
	ErrorCode           string
	Message             string
	RequestID           string
}
