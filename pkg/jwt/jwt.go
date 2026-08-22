package jwt

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/azmiagr/cakra-hackathon/entity"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Interface interface {
	CreateJWTToken(userID uuid.UUID, roleName string, sessionVersion int) (string, error)
	ValidateToken(tokenString string) (TokenIdentity, error)
	GetLoginUser(c *gin.Context) (*entity.User, error)
}

type jsonWebToken struct {
	SecretKey   string
	ExpiredTime time.Duration
}

type Claims struct {
	UserID         uuid.UUID
	IsAdmin        bool
	RoleName       string
	SessionVersion int
	jwt.RegisteredClaims
}

type TokenIdentity struct {
	UserID         uuid.UUID
	SessionVersion int
}

func Init() Interface {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	expiredTime, err := strconv.Atoi(os.Getenv("JWT_EXP_TIME"))
	if err != nil {
		log.Fatalf("error init jwt %v", err)
	}

	return &jsonWebToken{
		SecretKey:   secretKey,
		ExpiredTime: time.Duration(expiredTime) * time.Hour,
	}
}

func (j *jsonWebToken) CreateJWTToken(userID uuid.UUID, roleName string, sessionVersion int) (string, error) {
	claims := &Claims{
		UserID:         userID,
		RoleName:       roleName,
		SessionVersion: sessionVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ExpiredTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *jsonWebToken) ValidateToken(tokenString string) (TokenIdentity, error) {
	var claim Claims

	token, err := jwt.ParseWithClaims(tokenString, &claim, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return TokenIdentity{}, err
	}

	if !token.Valid {
		return TokenIdentity{}, errors.New("token is not valid")
	}

	return TokenIdentity{UserID: claim.UserID, SessionVersion: claim.SessionVersion}, nil
}

func (j *jsonWebToken) GetLoginUser(c *gin.Context) (*entity.User, error) {
	user, ok := c.Get("user")
	if !ok {
		return &entity.User{}, errors.New("failed to get user login")
	}

	return user.(*entity.User), nil
}
