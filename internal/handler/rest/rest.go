package rest

import (
	"fmt"
	"os"

	"github.com/azmiagr/cakra-hackathon/internal/service"
	"github.com/azmiagr/cakra-hackathon/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Rest struct {
	router     *gin.Engine
	service    *service.Service
	middleware middleware.Interface
}

func NewRest(service *service.Service, middleware middleware.Interface) *Rest {
	return &Rest{
		router:     gin.Default(),
		service:    service,
		middleware: middleware,
	}
}

func (r *Rest) MountEndpoint() {
	r.router.Use(r.middleware.Cors())
	r.router.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})
	baseUrl := r.router.Group("/api/v1")

	auth := baseUrl.Group("/auth")
	auth.POST("/login", r.Login)
	auth.POST("/logout", r.middleware.AuthenticateUser, r.Logout)
	auth.POST("/register", r.Register)
	auth.POST("/register/verify-otp", r.VerifyRegistrationOTP)
	auth.POST("/register/resend-otp", r.ResendRegistrationOTP)
	auth.POST("/register/password", r.SetRegistrationPassword)
	auth.POST("/forgot-password", r.RequestPasswordReset)
	auth.POST("/forgot-password/verify-otp", r.VerifyPasswordResetOTP)
	auth.POST("/forgot-password/resend-otp", r.ResendPasswordResetOTP)
	auth.POST("/forgot-password/password", r.SetPasswordReset)

	analysis := baseUrl.Group("/analysis")
	analysis.Use(r.middleware.AuthenticateUser)
	analysis.GET("/dashboard", r.GetDashboard)
	analysis.GET("/categories", r.GetCategories)
	analysis.GET("/history", r.GetAnalysisHistory)
	analysis.GET("/sessions/:sessionID", r.GetAnalysisSession)
	analysis.GET("/credit-account", r.GetCreditAccount)
	analysis.POST("/upload", r.UploadAnalysisXLSX)
	analysis.POST("/sessions/:uploadID", r.CreateAnalysisSession)

	internal := baseUrl.Group("/internal")
	internal.Use(r.middleware.AuthenticateAICallback)
	internal.POST("/analysis-sessions/:sessionID/result", r.CompleteAnalysisFromAI)
}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}
