package util

import (
	"store/internal/config"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTPayload struct {
	Email  string `json:"email"`
	UserId string `json:"uid"`
	jwt.RegisteredClaims
}

func GenerateAuthToken(email string, UserId string, config *config.ApplicationConfig) (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := jwt.MapClaims{
		"email": email,
		"uid":   UserId,
		"exp":   expirationTime.Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "bookstore-auth",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTSecretKey)

	if err != nil {
		return "", fmt.Errorf("could not create token: %v", err)
	}

	return tokenString, nil
}

func GetUserFromContext(c *gin.Context, config *config.ApplicationConfig) (*JWTPayload, error) {
	if val, exists := c.Get("userId"); exists {
		if userIdInt, ok := val.(int); ok {
			return &JWTPayload{UserId: fmt.Sprintf("%d", userIdInt)}, nil
		}
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid authorization format")
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return config.JWTSecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return &JWTPayload{
		UserId: claims["uid"].(string),
		Email:  claims["email"].(string),
	}, nil
}

