package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/azmiagr/cakra-hackathon/pkg/config"
)

type httpPredictor struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewHTTPPredictor(cfg config.AIConfig) Predictor {
	return &httpPredictor{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		client:  &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (p *httpPredictor) Predict(ctx context.Context, input PredictionInput) (PredictionOutput, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return PredictionOutput{}, &RequestError{
			Code: ErrorCodeInvalidResponse,
			err:  err,
		}
	}

	predictURL, err := url.JoinPath(p.baseURL, "predict")
	if err != nil {
		return PredictionOutput{}, &RequestError{
			Code: ErrorCodeInvalidResponse,
			err:  err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, predictURL, bytes.NewReader(body))
	if err != nil {
		return PredictionOutput{}, &RequestError{
			Code: ErrorCodeInvalidResponse,
			err:  err,
		}
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("X-API-Key", p.apiKey)
	}

	response, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return PredictionOutput{}, &RequestError{
				Code: ErrorCodeTimeout,
				err:  err,
			}
		}
		return PredictionOutput{}, &RequestError{
			Code: ErrorCodeUnavailable,
			err:  err,
		}
	}
	defer response.Body.Close()

	requestID := response.Header.Get("X-Request-ID")
	output, err := decodePredictionResponse(response, requestID)
	if err != nil {
		return PredictionOutput{}, err
	}
	log.Printf("AI prediction completed session_id=%s ai_request_id=%s status=%s", input.SessionID, output.RequestID, output.Status)
	return output, nil
}
