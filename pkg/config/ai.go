package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type AIConfig struct {
	CallbackSecret string
	BaseURL        string
	RequestTimeout time.Duration
	APIKey         string
}

func LoadAIConfig() (AIConfig, error) {
	secret := os.Getenv("AI_CALLBACK_SECRET")
	if secret == "" {
		return AIConfig{}, fmt.Errorf("AI_CALLBACK_SECRET is required")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("AI_BASE_URL")), "/")
	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return AIConfig{}, fmt.Errorf("AI_BASE_URL must be a valid HTTP or HTTPS URL")
	}

	timeoutSeconds, err := strconv.Atoi(os.Getenv("AI_REQUEST_TIMEOUT_SECONDS"))
	if err != nil || timeoutSeconds <= 0 {
		return AIConfig{}, fmt.Errorf("AI_REQUEST_TIMEOUT_SECONDS must be a positive integer")
	}

	return AIConfig{
		CallbackSecret: secret,
		BaseURL:        baseURL,
		RequestTimeout: time.Duration(timeoutSeconds) * time.Second,
		APIKey:         strings.TrimSpace(os.Getenv("AI_API_KEY")),
	}, nil
}
