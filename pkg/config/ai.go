package config

import (
	"fmt"
	"os"
)

type AIConfig struct{ CallbackSecret string }

func LoadAIConfig() (AIConfig, error) {
	secret := os.Getenv("AI_CALLBACK_SECRET")
	if secret == "" {
		return AIConfig{}, fmt.Errorf("AI_CALLBACK_SECRET is required")
	}
	return AIConfig{
		CallbackSecret: secret,
	}, nil
}
