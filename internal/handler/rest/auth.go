package rest

import (
	"net/http"

	"github.com/azmiagr/cakra-hackathon/model"
	"github.com/azmiagr/cakra-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

const registrationSessionHeader = "X-Session-Token"

func (r *Rest) Login(c *gin.Context) {
	var req model.LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "login berhasil", result)
}

func (r *Rest) Logout(c *gin.Context) {
	user, err := r.middleware.GetAuthenticatedUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "failed to get authenticated user", nil)
		return
	}

	err = r.service.AuthService.Logout(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "logout berhasil", nil)
}

func (r *Rest) Register(c *gin.Context) {
	var req model.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.Register(req)
	if result != nil {
		c.Header(registrationSessionHeader, result.SessionToken)
	}
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "OTP telah dikirim ke email", result)
}

func (r *Rest) VerifyRegistrationOTP(c *gin.Context) {
	var req model.VerifyRegistrationOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.VerifyRegistrationOTP(c.GetHeader(registrationSessionHeader), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header(registrationSessionHeader, result.SessionToken)
	response.Success(c, http.StatusOK, "OTP berhasil diverifikasi", nil)
}

func (r *Rest) ResendRegistrationOTP(c *gin.Context) {
	err := r.service.AuthService.ResendRegistrationOTP(c.GetHeader(registrationSessionHeader))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "OTP telah dikirim ulang", nil)
}

func (r *Rest) SetRegistrationPassword(c *gin.Context) {
	var req model.SetRegistrationPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.SetRegistrationPassword(c.GetHeader(registrationSessionHeader), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "akun berhasil diaktifkan", result)
}

func (r *Rest) RequestPasswordReset(c *gin.Context) {
	var req model.ForgotPasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.RequestPasswordReset(req)
	if result != nil {
		c.Header(registrationSessionHeader, result.SessionToken)
	}
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusAccepted, "Jika email terdaftar, instruksi reset telah dikirim", nil)
}

func (r *Rest) VerifyPasswordResetOTP(c *gin.Context) {
	var req model.VerifyPasswordResetOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	result, err := r.service.AuthService.VerifyPasswordResetOTP(c.GetHeader(registrationSessionHeader), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Header(registrationSessionHeader, result.SessionToken)
	response.Success(c, http.StatusOK, "OTP berhasil diverifikasi", nil)
}

func (r *Rest) ResendPasswordResetOTP(c *gin.Context) {
	err := r.service.AuthService.ResendPasswordResetOTP(c.GetHeader(registrationSessionHeader))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusAccepted, "Jika email terdaftar, instruksi reset telah dikirim", nil)
}

func (r *Rest) SetPasswordReset(c *gin.Context) {
	var req model.SetPasswordResetRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "request tidak valid", nil)
		return
	}

	err = r.service.AuthService.SetPasswordReset(c.GetHeader(registrationSessionHeader), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "password berhasil diubah, silakan masuk kembali", nil)
}
