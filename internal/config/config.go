package config

import (
	"errors"
	"os"
	"go.uber.org/zap"
	"github.com/joho/godotenv"
)

type ApplicationConfig struct {
	DSN                  string
	JWTSecretKey         []byte
	StripePublishableKey string
	StripeSecretKey      string
	StripeWebhookSecret  string
	InternalAPIKey       string
    MessageQueueURI      string
}


func InitConfig() (*ApplicationConfig, error) {
	config := ApplicationConfig{}
	err := godotenv.Load()
	if err != nil {
		zap.L().Error("Error loading .env file", zap.Error(err))
	}

	config.DSN = os.Getenv("dsn")
	config.JWTSecretKey = []byte(os.Getenv("JWTSecretKey"))
	config.StripePublishableKey = os.Getenv("StripePublishableKey")
	config.StripeSecretKey = os.Getenv("StripeSecretKey")
	config.StripeWebhookSecret = os.Getenv("StripeWebhookSecret")
	config.InternalAPIKey = os.Getenv("InternalAPIKey")
    config.MessageQueueURI = os.Getenv("MessageQueueURI")

	if config.DSN == "" || len(config.JWTSecretKey) == 0 || config.StripePublishableKey == "" || config.StripeSecretKey == "" || config.MessageQueueURI == "" {
		return nil, errors.New("DB DSN, JWTSecretKey, or Stripe keys, or MQURI not provided")
	}


	return &config, nil
}
