package orders

import (
    "store/internal/config"
    "store/internal/models"
    AuthRepo "store/internal/repositories/auth"
    "store/internal/repositories/orders"
    PaymentRepo "store/internal/repositories/payments"
    "context"
    "encoding/json"
    "fmt"
    "github.com/stripe/stripe-go/v74"
    "github.com/stripe/stripe-go/v74/checkout/session"
    "github.com/stripe/stripe-go/v74/refund"
    "github.com/stripe/stripe-go/v74/webhook"
    "github.com/stripe/stripe-go/v74/paymentmethod"
    "log"
    "math"
)

type OrderService interface {
    CreateOrder(ctx context.Context, userId int) (models.Order, error)
    CreateOrderItem(ctx context.Context, payload models.CreateOrderItemPayload, userId string) (models.OrderItem, error)
    GetUserOrders(ctx context.Context, userId string) ([]models.Order, error)
    CreateCheckoutSession(ctx context.Context, orderId int, userId int, successUrl, cancelUrl string) (string, error)
	RefundOrder(ctx context.Context, orderID int, items []models.RefundItem, reason string) error
    HandleWebhook(ctx context.Context, payload []byte, sigHeader string) error
}

type orderServiceImpl struct {
    repo        orders.OrderRepository
    authRepo    AuthRepo.AuthRepository
    paymentRepo PaymentRepo.PaymentRepo
    config *config.ApplicationConfig
}

func NewOrdersService(repo orders.OrderRepository, authRepo AuthRepo.AuthRepository, paymentRepo PaymentRepo.PaymentRepo, config *config.ApplicationConfig) OrderService {
    stripe.Key = config.StripeSecretKey
    return &orderServiceImpl{repo: repo, authRepo: authRepo, paymentRepo: paymentRepo, config: config}
}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, userId int) (models.Order, error) {
    return s.repo.CreateOrder(ctx, userId)
}

func (s *orderServiceImpl) CreateOrderItem(ctx context.Context, payload models.CreateOrderItemPayload, userId string) (models.OrderItem, error) {
    return s.repo.CreateOrderItem(ctx, payload, userId)
}

func (s *orderServiceImpl) GetUserOrders(ctx context.Context, userId string) ([]models.Order, error) {
    return s.repo.GetUserOrders(ctx, userId)
}

func (s *orderServiceImpl) CreateCheckoutSession(ctx context.Context, orderId int, userId int, successUrl, cancelUrl string) (string, error) {
    order, err := s.repo.GetOrderById(ctx, orderId)
    if err != nil {
        return "", err
    }

    if order.UserId != userId {
        return "", fmt.Errorf("unauthorized")
    }

    if order.Status == "paid" {
        return "", fmt.Errorf("order already paid")
    }

    u, err := s.authRepo.GetById(ctx, userId)
    if err != nil {
        return "", err
    }

    params := &stripe.CheckoutSessionParams{
        Customer:   stripe.String(u.StripeCustomerId),
        SuccessURL: stripe.String(successUrl),
        CancelURL:  stripe.String(cancelUrl),
        PaymentMethodTypes: stripe.StringSlice([]string{
            "card",
        }),
        Mode: stripe.String(string(stripe.CheckoutSessionModeSetup)),
        SetupIntentData: &stripe.CheckoutSessionSetupIntentDataParams{
            Metadata: map[string]string{
                "order_id": fmt.Sprintf("%d", orderId),
                "user_id":  fmt.Sprintf("%d", userId),
            },
        },
    }
    params.AddMetadata("order_id", fmt.Sprintf("%d", orderId))
    params.AddMetadata("user_id", fmt.Sprintf("%d", userId))

    if u.StripeCustomerId != "" {
        params.Customer = stripe.String(u.StripeCustomerId)
    } else {
        params.CustomerEmail = stripe.String(u.Email)
    }

    sess, err := session.New(params)
    if err != nil {
        return "", err
    }

    return sess.URL, nil
}

func (s *orderServiceImpl) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) error {
    /*
    Messages Names : 

        checkout.session.succeeded
        setup_intent.succeeded

    */

    // event, err := webhook.ConstructEvent(payload, sigHeader, config.AppConfig.StripeWebhookSecret,) -> verify signature
    event, err := webhook.ConstructEventWithOptions(
        payload,
        sigHeader,
        s.config.StripeWebhookSecret,
        webhook.ConstructEventOptions{
            IgnoreAPIVersionMismatch: true,
        },
    )
    if err != nil {
        return fmt.Errorf("failed to construct event: %v", err)
    }

    // Idempotency check
    processed, err := s.repo.IsEventProcessed(ctx, event.ID)
    if err != nil {
        return err
    }
    if processed {
        return nil
    }

    switch event.Type {
    case "checkout.session.completed":
        var sess stripe.CheckoutSession
        err := json.Unmarshal(event.Data.Raw, &sess)
        if err != nil { 
            log.Println("Error Unmarshaling json")
        }
        sessPrettyJSON, err := json.MarshalIndent(sess, "", "  ")
        if err == nil {
            fmt.Println("Pretty JSON:\n", string(sessPrettyJSON))
        }

        CustomerId := sess.Customer.ID
        PaymentIntentId := sess.PaymentIntent.ID
        var OrderId int
        OrderId, err = s.repo.GetCartId(ctx, CustomerId)
        fmt.Println("OrderId : ", OrderId)
        if err != nil { 
            log.Println("Error Fetching OrderId")
        }

        // update db ( ordered -> true & status -> paid )
        err = s.repo.UpdateOrderStatus(ctx, OrderId, "paid") 
        if err != nil { 
            log.Println("Error Updating Order Status")
        }
        // update payment_intent_id / checkout_session_id to order
        err = s.repo.UpdateOrderPaymentIntentId(ctx, OrderId, PaymentIntentId) 
        if err != nil { 
            log.Println("Error Updating Order Status")
        }

    case "setup_intent.succeeded":
        log.Println("setup_intent.succeeded")

        // Unmarshal into SetupIntent
        var si stripe.SetupIntent
        if err := json.Unmarshal(event.Data.Raw, &si); err != nil {
            return err
        }

        fmt.Println("setupIndentObj:", si)

        // Get the PaymentMethod ID from the SetupIntent
        if si.PaymentMethod == nil {
            return fmt.Errorf("SetupIntent has no PaymentMethod")
        }
        pmID := si.PaymentMethod.ID

        // Fetch full PaymentMethod from Stripe
        pm, err := paymentmethod.Get(pmID, nil)
        if err != nil {
            return err
        }

        // Now safe to access Card and BillingDetails
        var Brand, last4, cardHolder string
        var expMonth, expYear int

        if pm.Card != nil {
            Brand = string(pm.Card.Brand)
            last4 = pm.Card.Last4
            expMonth = int(pm.Card.ExpMonth)
            expYear = int(pm.Card.ExpYear)
        }

        if pm.BillingDetails != nil {
            cardHolder = pm.BillingDetails.Name
        }

        customerID := pm.Customer.ID

        // Lookup user and save
        UserObj, err := s.repo.GetUserByStripeCustomerID(ctx, customerID)
        if err != nil {
            return err
        }

        fmt.Println("userId : ", UserObj.ID)
        fmt.Println("stripeCustomerId : ", UserObj.StripeCustomerId)

        return s.repo.SavePaymentMethod(ctx, models.PaymentMethods{
            UserID:           int(UserObj.ID),
            StripeCustomerID: UserObj.StripeCustomerId,
            PaymentMethodID:  pmID,
            Brand:            Brand,
            Last4:            last4,
            ExpiryMonth:      expMonth,
            ExpiryYear:       expYear,
            CardHolder:       cardHolder,
        })

    }
    // add event to stripe processed events table.
    return s.repo.MarkEventProcessed(ctx, event.ID)
}

 func (s *orderServiceImpl) RefundOrder(ctx context.Context, orderID int, items []models.RefundItem, reason string) error{
	order, err := s.repo.GetOrderById(ctx, orderID)
	if err != nil {
		return err
	}

	if order.PaymentIntentId == "" {
		return fmt.Errorf("order has no payment intent")
	}

	// Calculate total refund amount based on selected items
	var refundAmount float64
	for _, item := range items {
		orderItem, err := s.repo.GetOrderItemById(ctx, item.ID)
		if err != nil {
			return err
		}
		if item.Quantity > orderItem.Quantity {
			return fmt.Errorf("refund quantity exceeds purchased quantity for item %d", item.ID)
		}
		refundAmount += float64(item.Quantity) * orderItem.Price
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(order.PaymentIntentId),
		Amount:        stripe.Int64(int64(math.Round(refundAmount * 100))),
	}
	stRefund, err := refund.New(params)
	if err != nil {
		return err
	}

	// Record refund in DB
	dbRefund := models.Refund{
		OrderId:        order.Id,
		StripeRefundId: stRefund.ID,
		Amount:         refundAmount,
		Reason:         reason,
		Status:         string(stRefund.Status),
	}

	if err := s.repo.CreateRefund(ctx, dbRefund); err != nil {
		return err
	}

	// Update order status if fully refunded
	totalRefunded, _ := s.repo.GetTotalRefundedAmount(ctx, order.Id)
	if totalRefunded >= order.TotalAmount {
		return s.repo.UpdateOrderStatus(ctx, order.Id, "refunded")
	} else {
		return s.repo.UpdateOrderStatus(ctx, order.Id, "partially_refunded")
	}
}

