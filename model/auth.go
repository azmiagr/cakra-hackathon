package model

import "time"

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=3,max=150"`
	Email    string `json:"email" binding:"required,email,max=100"`
}

type VerifyRegistrationOTPRequest struct {
	OTP string `json:"otp" binding:"required,len=6,numeric"`
}

type SetRegistrationPasswordRequest struct {
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type RegistrationResult struct {
	SessionToken string    `json:"session_token"`
	OTPExpiresAt time.Time `json:"otp_expires_at"`
}

type CompleteRegistrationResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// model/auth.go
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=100"`
}

type VerifyPasswordResetOTPRequest struct {
	OTP string `json:"otp" binding:"required,len=6,numeric"`
}

type SetPasswordResetRequest struct {
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type PasswordResetResult struct {
	SessionToken string
}
