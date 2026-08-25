package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxResponseBodyBytes = 1 << 20

type predictResponse struct {
	Status              string    `json:"status"`
	DemandCategory      string    `json:"demand_category"`
	ADIValue            *float64  `json:"adi_value"`
	CVSquaredValue      *float64  `json:"cv_squared_value"`
	ForecastHorizonDays int       `json:"forecast_horizon_days"`
	ForecastP50         []float64 `json:"forecast_p50"`
	ForecastP90         []float64 `json:"forecast_p90"`
	ReorderPoint        int       `json:"reorder_point"`
	ReorderQuantity     int       `json:"reorder_quantity"`
	RiskLabel           string    `json:"risk_label"`
	RiskReason          string    `json:"risk_reason"`
	ExplanationText     string    `json:"explanation_text"`
	ErrorCode           string    `json:"error_code"`
	Message             string    `json:"message"`
}

func decodePredictionResponse(response *http.Response, requestID string) (PredictionOutput, error) {
	var payload predictResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBodyBytes))
	if err := decoder.Decode(&payload); err != nil {
		return PredictionOutput{}, &RequestError{Code: ErrorCodeInvalidResponse, err: err}
	}

	output := PredictionOutput{
		Status:              payload.Status,
		DemandCategory:      payload.DemandCategory,
		ADIValue:            payload.ADIValue,
		CVSquaredValue:      payload.CVSquaredValue,
		ForecastHorizonDays: payload.ForecastHorizonDays,
		ForecastP50:         payload.ForecastP50,
		ForecastP90:         payload.ForecastP90,
		ReorderPoint:        payload.ReorderPoint,
		ReorderQuantity:     payload.ReorderQuantity,
		RiskLabel:           payload.RiskLabel,
		RiskReason:          payload.RiskReason,
		ExplanationText:     payload.ExplanationText,
		ErrorCode:           payload.ErrorCode,
		Message:             payload.Message,
		RequestID:           requestID,
	}

	switch response.StatusCode {
	case http.StatusOK:
		if !validSuccessOutput(output) {
			return PredictionOutput{}, &RequestError{Code: ErrorCodeInvalidResponse, err: fmt.Errorf("invalid AI success response")}
		}
	case http.StatusUnprocessableEntity:
		if output.Status != "INSUFFICIENT_DATA" && output.Status != "VALIDATION_ERROR" {
			return PredictionOutput{}, &RequestError{Code: ErrorCodeInvalidResponse, err: fmt.Errorf("invalid AI validation response")}
		}
	case http.StatusInternalServerError:
		if output.Status != "FAILED" {
			return PredictionOutput{}, &RequestError{Code: ErrorCodeInvalidResponse, err: fmt.Errorf("invalid AI failure response")}
		}
	default:
		return PredictionOutput{}, &RequestError{Code: ErrorCodeUnexpectedResponse, err: fmt.Errorf("AI returned HTTP %d", response.StatusCode)}
	}

	return output, nil
}

func validSuccessOutput(output PredictionOutput) bool {
	if output.Status != "SUCCESS" || output.DemandCategory == "" || output.RiskLabel == "" || output.RiskReason == "" || output.ExplanationText == "" || output.ForecastHorizonDays <= 0 || len(output.ForecastP50) != output.ForecastHorizonDays || len(output.ForecastP90) != output.ForecastHorizonDays {
		return false
	}
	for index, p50 := range output.ForecastP50 {
		if output.ForecastP90[index] < p50 {
			return false
		}
	}
	return true
}
