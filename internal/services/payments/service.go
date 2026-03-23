package payment

import (
	"store/internal/config"
	"store/internal/models"
	AuthRepo "store/internal/repositories/auth"
	PaymentRepo "store/internal/repositories/payments"
    OrderRepo "store/internal/repositories/orders"
	"context"
	"fmt"
    "log"
    "errors"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/setupintent"
    "github.com/stripe/stripe-go/v74/checkout/session"
)

type PaymentService interface {
	CreateSetupIntent(ctx context.Context, userId int) (string, error)
	AddPaymentMethod(ctx context.Context, userId int, stripePmId string) (models.PaymentMethods, error)
	DeletePaymentMethod(ctx context.Context, userId int, stripePmId string) error
	ListPaymentMethods(ctx context.Context, userId int) ([]models.PaymentMethods, error)
	GetOrCreateStripeCustomer(ctx context.Context, userId int) (string, error)
    CreateCheckoutSession(ctx context.Context, OrderId int) (string, error) 
}

type paymentService struct {
	repo     PaymentRepo.PaymentRepo
	authRepo AuthRepo.AuthRepository
    orderRepo OrderRepo.OrderRepository
    config *config.ApplicationConfig
}



func NewPaymentService(repo PaymentRepo.PaymentRepo, authRepo AuthRepo.AuthRepository, orderRepo OrderRepo.OrderRepository, config *config.ApplicationConfig) PaymentService {
	stripe.Key = config.StripeSecretKey
    return &paymentService{repo: repo, authRepo: authRepo, orderRepo: orderRepo, config: config}
}


func (s *paymentService) CreateCheckoutSession(ctx context.Context, OrderID int) (string, error) {
    orderObj, err := s.orderRepo.GetOrderById(ctx, OrderID)
    if err != nil { 
        log.Println("Invalid Order Id")
    }
    orderStatus := orderObj.Status
    if orderStatus == "paid" { 
        return "", errors.New("Order Already Paid")

    }

    items, err := s.orderRepo.GetOrderItems(ctx, OrderID)
    if err != nil {
        return "", err
    }

    userObj, err := s.authRepo.GetById(ctx, orderObj.UserId)
    if err != nil {
        return "", err
    }

    customerID := userObj.StripeCustomerId

    var stripeLineItems []*stripe.CheckoutSessionLineItemParams
    for _, item := range items {
        stripeLineItems = append(stripeLineItems, &stripe.CheckoutSessionLineItemParams{
            PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                Currency: stripe.String("usd"),
                ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                    Name: stripe.String(item.Title),
                },
                UnitAmount: stripe.Int64(int64(item.Price * 100)),
            },
            Quantity: stripe.Int64(int64(item.Quantity)),
        })
    }

    params := &stripe.CheckoutSessionParams{
        Customer: stripe.String(customerID),
        PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
        LineItems:          stripeLineItems,
        Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
        SuccessURL:         stripe.String("https://example.com/success?id="),
        CancelURL:          stripe.String("https://example.com/cancel"),
    }

    sess, err := session.New(params)
    if err != nil {
        return "", fmt.Errorf("stripe session creation: %w", err)
    }

    return sess.URL, nil
}




func (s *paymentService) GetOrCreateStripeCustomer(ctx context.Context, userId int) (string, error) {
	u, err := s.authRepo.GetById(ctx, userId)
	if err != nil {
		return "", err
	}

	if u.StripeCustomerId != "" {
		return u.StripeCustomerId, nil
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(u.Email),
	}
	params.AddMetadata("user_id", fmt.Sprintf("%d", userId))

	cus, err := customer.New(params)
	if err != nil {
		return "", err
	}

	err = s.authRepo.UpdateStripeCustomerId(ctx, userId, cus.ID)
	if err != nil {
		return "", err
	}

	return cus.ID, nil
}

func (s *paymentService) CreateSetupIntent(ctx context.Context, userId int) (string, error) {
	cusID, err := s.GetOrCreateStripeCustomer(ctx, userId)
	if err != nil {
		return "", err
	}

	params := &stripe.SetupIntentParams{
		Customer: stripe.String(cusID),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
	}
	si, err := setupintent.New(params)
	if err != nil {
		return "", err
	}

	return si.ClientSecret, nil
}

func (s *paymentService) AddPaymentMethod(ctx context.Context, userId int, stripePmId string) (models.PaymentMethods, error) {
	cusID, err := s.GetOrCreateStripeCustomer(ctx, userId)
	if err != nil {
		return models.PaymentMethods{}, err
	}

	// Attach payment method to customer if not already attached
	pm, err := paymentmethod.Get(stripePmId, nil)
	if err != nil {
		return models.PaymentMethods{}, err
	}

	if pm.Customer == nil || pm.Customer.ID != cusID {
		params := &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(cusID),
		}
		_, err = paymentmethod.Attach(stripePmId, params)
		if err != nil {
			return models.PaymentMethods{}, err
		}
	}

    dbPm := models.PaymentMethods{
        UserID:           userId,
        StripeCustomerID: cusID,           
        PaymentMethodID:  stripePmId,
        Brand:            string(pm.Card.Brand),
        Last4:            pm.Card.Last4,   
        CardHolder:       pm.BillingDetails.Name,
        ExpiryMonth:      int(pm.Card.ExpMonth),
        ExpiryYear:       int(pm.Card.ExpYear),
    }

	return s.repo.Save(ctx, dbPm)
}

func (s *paymentService) DeletePaymentMethod(ctx context.Context, userId int, stripePmId string) error {
	// Detach from Stripe
	_, err := paymentmethod.Detach(stripePmId, nil)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, userId, stripePmId)
}

func (s *paymentService) ListPaymentMethods(ctx context.Context, userId int) ([]models.PaymentMethods, error) {
	return s.repo.GetAll(ctx, userId)
}

