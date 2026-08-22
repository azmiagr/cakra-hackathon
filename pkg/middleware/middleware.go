package middleware

import (
	"crypto/subtle"
	"github.com/azmiagr/cakra-hackathon/internal/service"
	"github.com/azmiagr/cakra-hackathon/pkg/config"
	"github.com/azmiagr/cakra-hackathon/pkg/jwt"
	"github.com/azmiagr/cakra-hackathon/pkg/response"
	"github.com/gin-gonic/gin"
)

type Interface interface {
	Cors() gin.HandlerFunc
	AuthenticateUser(c *gin.Context)
	AuthenticateAICallback(c *gin.Context)
}

type middleware struct {
	service  *service.Service
	jwtAuth  jwt.Interface
	aiConfig config.AIConfig
}

func Init(service *service.Service, jwtAuth jwt.Interface, aiConfig config.AIConfig) Interface {
	return &middleware{
		service:  service,
		jwtAuth:  jwtAuth,
		aiConfig: aiConfig,
	}
}

func (m *middleware) AuthenticateAICallback(c *gin.Context) {
	provided := c.GetHeader("X-AI-Callback-Secret")
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(m.aiConfig.CallbackSecret)) != 1 {
		response.Error(c, 401, "invalid AI callback credentials", nil)
		c.Abort()
		return
	}
	c.Next()
}
