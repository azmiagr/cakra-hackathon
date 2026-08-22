package helper

import (
	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/azmiagr/cakra-hackathon/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUserIDParam(c *gin.Context) (uuid.UUID, error) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		return uuid.Nil, errors.BadRequest("invalid user id")
	}

	return userID, nil
}

func ParseUUIDParam(c *gin.Context, paramName string, message string) (uuid.UUID, error) {
	value, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		return uuid.Nil, errors.BadRequest(message)
	}

	return value, nil
}

func GetAuthenticatedUser(c *gin.Context) (*entity.User, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, errors.Unauthorized("unauthorized")
	}

	user, ok := userValue.(*entity.User)
	if !ok {
		return nil, errors.Unauthorized("unauthorized")
	}

	return user, nil
}
