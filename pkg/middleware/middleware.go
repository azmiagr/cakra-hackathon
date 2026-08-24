package middleware

import (
	"crypto/subtle"
	"errors"

	"github.com/azmiagr/cakra-hackathon/entity"
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
	GetAuthenticatedUser(c *gin.Context) (*entity.User, error)
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

func (m *middleware) GetAuthenticatedUser(c *gin.Context) (*entity.User, error) {
	value, ok := c.Get("user")
	if !ok {
		return nil, errors.New("authenticated user is missing from context")
	}

	user, ok := value.(*entity.User)
	if !ok {
		return nil, errors.New("authenticated user has an invalid context type")
	}

	return user, nil
}
