package orders

import (
	"store/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userId int) (models.Order, error)
	CreateOrderItem(ctx context.Context, payload models.CreateOrderItemPayload, userId string) (models.OrderItem, error)
	GetUserOrders(ctx context.Context, userId string) ([]models.Order, error)
	GetOrderById(ctx context.Context, orderId int) (models.Order, error)
	GetOrderItems(ctx context.Context, orderId int) ([]models.OrderItemsAPIResponse, error)
	UpdateOrderStatus(ctx context.Context, orderId int, status string) error
	CreateRefund(ctx context.Context, refund models.Refund) error
	IsEventProcessed(ctx context.Context, eventId string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventId string) error
    GetUserByStripeCustomerID(ctx context.Context, stripeCustomerId string) (models.User, error)
    SavePaymentMethod(ctx context.Context, params models.PaymentMethods) error
    GetCartId(ctx context.Context, stripeCustomerId string) (int, error)
    UpdateOrderPaymentIntentId(ctx context.Context, orderId int, paymentIntentId string) error
    GetOrderItemById(ctx context.Context, ItemId int) (models.OrderItem, error) 
    GetTotalRefundedAmount(ctx context.Context, orderItem int) (float64, error)
}

type PostgresOrderRepo struct {
	db *sql.DB
}

func NewOrdersRepo(db *sql.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) GetTotalRefundedAmount(ctx context.Context, orderId int) (float64, error) {
	var total float64

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM refunds
		WHERE order_id = $1
		AND status IN ('succeeded', 'pending');
	`

	err := r.db.QueryRowContext(ctx, query, orderId).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *PostgresOrderRepo) GetOrderItemById(ctx context.Context, itemId int) (models.OrderItem, error) {
	var item models.OrderItem

	query := `
		SELECT 
			id,
			order_id,
			book_id,
			quantity,
			price
		FROM order_items
		WHERE id = $1;
	`

	err := r.db.QueryRowContext(ctx, query, itemId).Scan(
		&item.Id,
		&item.OrderId,
		&item.BookId,
		&item.Quantity,
		&item.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.OrderItem{}, fmt.Errorf("order item not found")
		}
		return models.OrderItem{}, err
	}

	return item, nil
}


func (r *PostgresOrderRepo) UpdateOrderPaymentIntentId(ctx context.Context, orderId int, paymentIntentId string) error {
    query := `
        UPDATE orders
        SET payment_intent_id = $1,
            updated_at = NOW()
        WHERE id = $2;
    `
    _, err := r.db.ExecContext(ctx, query, paymentIntentId, orderId)
    if err != nil {
        return err
    }

    return nil
}


func (r *PostgresOrderRepo) GetCartId(ctx context.Context, stripeCustomerId string) (int, error) { 
    var OrderId int
    query := `
    select o.id 
    from orders o
    join users  u
    on o.user_id = u.id
    where  
    o.ordered = false
    and stripe_customer_id = $1
    limit 1;
    `
    err := r.db.QueryRowContext(ctx, query, stripeCustomerId).
        Scan(&OrderId)

    if err != nil {
        return 0, err
    }

    return OrderId, nil


}


func (r *PostgresOrderRepo) SavePaymentMethod(ctx context.Context, p models.PaymentMethods) error {
    query := `
        INSERT INTO user_payment_methods (
            user_id,
            stripe_customer_id,
            payment_method_id,
            brand,
            last4,
            card_holder,
            expiry_month,
            expiry_year
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (payment_method_id) DO NOTHING;
    `

    result, err := r.db.ExecContext(ctx, query,
        p.UserID,
        p.StripeCustomerID,
        p.PaymentMethodID,
        p.Brand,
        p.Last4,
        p.CardHolder,
        p.ExpiryMonth,
        p.ExpiryYear,
    )
    if err != nil {
        return fmt.Errorf("failed to save payment method for user %d: %w", p.UserID, err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        fmt.Printf("Payment method %s already exists, skipping.\n", p.PaymentMethodID)
    }

    return nil
}




func (r *PostgresOrderRepo) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerId string) (models.User, error) {
    var user models.User

    query := `
        SELECT id, stripe_customer_id, email
        FROM users
        WHERE stripe_customer_id = $1
        LIMIT 1;
    `

    err := r.db.QueryRowContext(ctx, query, stripeCustomerId).
        Scan(&user.ID, &user.StripeCustomerId, &user.Email)

    if err != nil {
        return models.User{}, err
    }
    fmt.Println("=============================GetUserByStripeCustomerID()================", user)

    return user, nil
}




func (r *PostgresOrderRepo) CreateOrder(ctx context.Context, userId int) (models.Order, error) {
	var order models.Order
	query := "SELECT id, ordered, user_id, status, total_amount, payment_intent_id, created_at, updated_at FROM orders WHERE ordered = false AND user_id = $1 LIMIT 1"
	row := r.db.QueryRowContext(ctx, query, userId)

	err := row.Scan(&order.Id, &order.Ordered, &order.UserId, &order.Status, &order.TotalAmount, &order.PaymentIntentId, &order.CreatedAt, &order.UpdatedAt)
	if err != nil && err != sql.ErrNoRows {
		return models.Order{}, err
	}

	if err == nil {
		return order, nil
	}

	insertQuery := `
		INSERT INTO orders (ordered, user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		RETURNING id, ordered, user_id, status, total_amount, created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, insertQuery, false, userId, "pending", 0.00).Scan(
		&order.Id,
		&order.Ordered,
		&order.UserId,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *PostgresOrderRepo) CreateOrderItem(ctx context.Context, payload models.CreateOrderItemPayload, userId string) (models.OrderItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.OrderItem{}, err
	}
	defer tx.Rollback()

	var orderID int
	var currentTotal float64

	getOrderQuery := `
		SELECT id, total_amount
		FROM orders 
		WHERE user_id = $1 AND ordered = false 
		LIMIT 1`

	err = tx.QueryRowContext(ctx, getOrderQuery, userId).Scan(&orderID, &currentTotal)

	if err != nil {
		if err == sql.ErrNoRows {
			createOrderQuery := `
				INSERT INTO orders (user_id, ordered, status, total_amount)
				VALUES ($1, false, 'pending', 0.00)
				RETURNING id`
			err = tx.QueryRowContext(ctx, createOrderQuery, userId).Scan(&orderID)
			if err != nil {
				return models.OrderItem{}, err
			}
		} else {
			return models.OrderItem{}, err
		}
	}

	var price float64
	priceQuery := `SELECT price FROM books WHERE id = $1`
	err = tx.QueryRowContext(ctx, priceQuery, payload.BookId).Scan(&price)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.OrderItem{}, fmt.Errorf("book not found")
		}
		return models.OrderItem{}, err
	}

	var itemID int
	var quantity int
	checkItemQuery := `SELECT id, quantity FROM order_items WHERE order_id = $1 AND book_id = $2`
	err = tx.QueryRowContext(ctx, checkItemQuery, orderID, payload.BookId).Scan(&itemID, &quantity)

	var orderItem models.OrderItem
	if err != nil {
		if err == sql.ErrNoRows {
			insertQuery := `
				INSERT INTO order_items (order_id, book_id, quantity, price)
				VALUES ($1, $2, $3, $4)
				RETURNING id`
			err = tx.QueryRowContext(ctx, insertQuery, orderID, payload.BookId, 1, price).Scan(&orderItem.Id)
			if err != nil {
				return models.OrderItem{}, err
			}
			orderItem.Quantity = 1
		} else {
			return models.OrderItem{}, err
		}
	} else {
		updateQuery := `
			UPDATE order_items
			SET quantity = quantity + 1
			WHERE id = $1
			RETURNING quantity`
		err = tx.QueryRowContext(ctx, updateQuery, itemID).Scan(&quantity)
		if err != nil {
			return models.OrderItem{}, err
		}
		orderItem.Id = itemID
		orderItem.Quantity = quantity
	}

	// Update order total
	_, err = tx.ExecContext(ctx, "UPDATE orders SET total_amount = total_amount + $1, updated_at = now() WHERE id = $2", price, orderID)
	if err != nil {
		return models.OrderItem{}, err
	}

	orderItem.OrderId = orderID
	orderItem.BookId = payload.BookId
	orderItem.Price = price

	if err := tx.Commit(); err != nil {
		return models.OrderItem{}, err
	}

	return orderItem, nil
}

func (r *PostgresOrderRepo) GetUserOrders(ctx context.Context, userId string) ([]models.Order, error) {
	query := `
		SELECT id, ordered, user_id, status, total_amount, payment_intent_id, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.Id, &order.Ordered, &order.UserId, &order.Status, &order.TotalAmount, &order.PaymentIntentId, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PostgresOrderRepo) GetOrderById(ctx context.Context, orderId int) (models.Order, error) {
	var order models.Order
    fmt.Println("orderId repo : ", orderId)
	query := "SELECT id, ordered, user_id, status, total_amount, payment_intent_id, created_at, updated_at FROM orders WHERE id = $1"
	err := r.db.QueryRowContext(ctx, query, orderId).Scan(&order.Id, &order.Ordered, &order.UserId, &order.Status, &order.TotalAmount, &order.PaymentIntentId, &order.CreatedAt, &order.UpdatedAt)
	return order, err
}

func (r *PostgresOrderRepo) GetOrderItems(ctx context.Context, orderId int) ([]models.OrderItemsAPIResponse, error) {
    query := `
        select 
            oi.id as "ItemId",
            b.title as "Title",
            oi.price as "Price",
            oi.quantity as "Quantity"
        from 
            order_items oi
            join books b
            on b.id = oi.book_id
        where order_id = $1;
`
	rows, err := r.db.QueryContext(ctx, query, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItemsAPIResponse
	for rows.Next() {
		var item models.OrderItemsAPIResponse
		if err := rows.Scan(&item.ItemId, &item.Title, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *PostgresOrderRepo) UpdateOrderStatus(ctx context.Context, orderId int, status string) error {
	ordered := false
	if status == "paid" {
		ordered = true
	}
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET status = $1, ordered = $2, updated_at = now() WHERE id = $3", status, ordered, orderId)
	return err
}


func (r *PostgresOrderRepo) CreateRefund(ctx context.Context, refund models.Refund) error {
	query := `INSERT INTO refunds (order_id, stripe_refund_id, amount, reason, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, refund.OrderId, refund.StripeRefundId, refund.Amount, refund.Reason, refund.Status)
	return err
}

func (r *PostgresOrderRepo) IsEventProcessed(ctx context.Context, eventId string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM stripe_processed_events WHERE event_id = $1)"
	err := r.db.QueryRowContext(ctx, query, eventId).Scan(&exists)
	return exists, err
}

func (r *PostgresOrderRepo) MarkEventProcessed(ctx context.Context, eventId string) error {
	query := "INSERT INTO stripe_processed_events (event_id) VALUES ($1)"
	_, err := r.db.ExecContext(ctx, query, eventId)
	return err
}

