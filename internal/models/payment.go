package models


import (
    "time"
)


type PaymentMethods struct {
    ID               int       `json:"id"`
    UserID           int       `json:"userId"`
    StripeCustomerID string    `json:"stripeCustomerId"`
    PaymentMethodID  string    `json:"paymentMethodId"`
    Brand            string    `json:"brand"`
    Last4            string    `json:"last4"`
    ExpiryMonth      int       `json:"expiryMonth"`
    ExpiryYear       int       `json:"expiryYear"`
    CardHolder       string    `json:"cardHolder"`
    CreatedAt        time.Time `json:"createdAt"`
    UpdatedAt        time.Time `json:"updatedAt"`
}

type SetupIntentResponse struct {
	ClientSecret string `json:"client_secret"`
}

