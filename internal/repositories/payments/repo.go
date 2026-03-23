package payments

import (
	"store/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type PaymentRepo interface {
	Save(ctx context.Context, pm models.PaymentMethods) (models.PaymentMethods, error)
	Delete(ctx context.Context, userId int, stripePmId string) error
	GetAll(ctx context.Context, userId int) ([]models.PaymentMethods, error)
}

type paymentRepo struct {
	db *sql.DB
}

func NewPaymentsRepo(db *sql.DB) PaymentRepo {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Save(ctx context.Context, pm models.PaymentMethods) (models.PaymentMethods, error) {
	query := `
        INSERT INTO user_payment_methods 
        (user_id, stripe_customer_id, payment_method_id, brand, last4, card_holder, expiry_month, expiry_year, created_at, updated_at)
        VALUES (6,'cus',$3,$4,$5,$6,$7,$8, now(), now())
        ON CONFLICT (payment_method_id) 
        DO UPDATE SET updated_at = now()
        RETURNING id, created_at, updated_at
    `
	err := r.db.QueryRowContext(ctx, query,
		pm.UserID,
		pm.StripeCustomerID,
		pm.PaymentMethodID,
		pm.Brand,
		pm.Last4,
		pm.CardHolder,
		pm.ExpiryMonth,
		pm.ExpiryYear,
	).Scan(&pm.ID, &pm.CreatedAt, &pm.UpdatedAt)

	if err != nil {
		return models.PaymentMethods{}, err
	}

	return pm, nil
}

// Delete removes a payment method for a user
func (r *paymentRepo) Delete(ctx context.Context, userId int, stripePmId string) error {
	query := `
        DELETE FROM user_payment_methods 
        WHERE payment_method_id=$1 AND user_id=$2
    `
	res, err := r.db.ExecContext(ctx, query, stripePmId, userId)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payment method not found")
	}
	return nil
}

// GetAll fetches all payment methods for a user
func (r *paymentRepo) GetAll(ctx context.Context, userId int) ([]models.PaymentMethods, error) {
	query := `
        SELECT id, user_id, stripe_customer_id, payment_method_id, brand, last4, card_holder, expiry_month, expiry_year, created_at, updated_at
        FROM user_payment_methods
        WHERE user_id=$1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []models.PaymentMethods
	for rows.Next() {
		var pm models.PaymentMethods
		if err := rows.Scan(
			&pm.ID,
			&pm.UserID,
			&pm.StripeCustomerID,
			&pm.PaymentMethodID,
			&pm.Brand,
			&pm.Last4,
			&pm.CardHolder,
			&pm.ExpiryMonth,
			&pm.ExpiryYear,
			&pm.CreatedAt,
			&pm.UpdatedAt,
		); err != nil {
			return nil, err
		}
		methods = append(methods, pm)
	}

	return methods, nil
}


