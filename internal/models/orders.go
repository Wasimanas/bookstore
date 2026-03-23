package models

import (
	"time"
)

type Order struct {
	Id              int       `json:"id"`
	Ordered         bool      `json:"ordered"`
	UserId          int       `json:"user_id"`
	Status          string    `json:"status"`
	TotalAmount     float64   `json:"total_amount"`
	PaymentIntentId string    `json:"payment_intent_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OrderItem struct {
	Id       int     `json:"id"`
	OrderId  int     `json:"order_id"`
	BookId   int     `json:"book_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type OrderItemsAPIResponse struct { 
    ItemId int      `json:"item_id"`
    Title string `json:"title"`
    Price float64   `json:"price"`
    Quantity int    `json:"quantity"`
}


type CreateOrderPayload struct {
	UserId int `json:"user_id"`
}

type CreateOrderItemPayload struct {
	BookId int `json:"book_id"`
}

type Refund struct {
	Id             int       `json:"id"`
	OrderId        int       `json:"order_id"`
	StripeRefundId string    `json:"stripe_refund_id"`
	Amount         float64   `json:"amount"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type StripeProcessedEvent struct {
	EventId     string    `json:"event_id"`
	ProcessedAt time.Time `json:"processed_at"`
}

type OrderStatus string

const (
    Pending   OrderStatus = "pending"
    Paid      OrderStatus = "paid"
    Shipped   OrderStatus = "shipped"
    Cancelled OrderStatus = "cancelled"
)

type RefundItem struct {
	ID       int
	Quantity int
}
