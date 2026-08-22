package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type RegistrationConfig struct {
	OTPTTL         time.Duration
	SessionTTL     time.Duration
	PasswordTTL    time.Duration
	ResendCooldown time.Duration
	MaxOTPAttempts int
}

func LoadRegistrationConfig() (RegistrationConfig, error) {
	otpMinutes, err := envPositiveInt("OTP_TTL_MINUTES", 5)
	if err != nil {
		return RegistrationConfig{}, err
	}

	sessionMinutes, err := envPositiveInt("REGISTRATION_SESSION_TTL_MINUTES", 15)
	if err != nil {
		return RegistrationConfig{}, err
	}

	passwordMinutes, err := envPositiveInt("PASSWORD_SESSION_TTL_MINUTES", 15)
	if err != nil {
		return RegistrationConfig{}, err
	}

	cooldownMinutes, err := envPositiveInt("OTP_RESEND_COOLDOWN_MINUTES", 5)
	if err != nil {
		return RegistrationConfig{}, err
	}

	maxAttempts, err := envPositiveInt("OTP_MAX_ATTEMPTS", 5)
	if err != nil {
		return RegistrationConfig{}, err
	}

	return RegistrationConfig{
		OTPTTL:         time.Duration(otpMinutes) * time.Minute,
		SessionTTL:     time.Duration(sessionMinutes) * time.Minute,
		PasswordTTL:    time.Duration(passwordMinutes) * time.Minute,
		ResendCooldown: time.Duration(cooldownMinutes) * time.Minute,
		MaxOTPAttempts: maxAttempts,
	}, nil
}

func envPositiveInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return value, nil
}
