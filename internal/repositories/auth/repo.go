package auth 


import (
	"store/internal/models"
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	Register(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, email, password string) (models.User, error)
	GetById(ctx context.Context, id int) (models.User, error)
	UpdateStripeCustomerId(ctx context.Context, userId int, stripeCustomerId string) error
}


type PostgresAuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) AuthRepository {
	return &PostgresAuthRepo{db: db}
}


func (r *PostgresAuthRepo) Register(ctx context.Context, user models.User) (models.User, error) {
	query := `INSERT INTO users (first_name, last_name, email, password, created_at, updated_at)
              VALUES ($1, $2, $3, $4, NOW(), NOW())
              RETURNING id, first_name, email, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query, user.FirstName, user.LastName, user.Email, user.Password)
	var u models.User
	u.Password = ""
	if err := row.Scan(&u.ID, &u.FirstName, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return models.User{}, err
	}

	return u, nil
}

func (r *PostgresAuthRepo) Login(ctx context.Context, email, password string) (models.User, error) {
	query := `SELECT id, first_name, email, password, stripe_customer_id, created_at, updated_at FROM users WHERE email = $1`
	var u models.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.FirstName, &u.Email, &u.Password, &u.StripeCustomerId, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid credentials")
	}
	u.Password = ""
	return u, nil
}


func (r *PostgresAuthRepo) GetById(ctx context.Context, id int) (models.User, error) {
	query := `SELECT id, first_name, email, stripe_customer_id, created_at, updated_at FROM users WHERE id = $1`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.FirstName, &user.Email, &user.StripeCustomerId, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}


func (r *PostgresAuthRepo) UpdateStripeCustomerId(ctx context.Context, userId int, stripeCustomerId string) error {
	query := "UPDATE users SET stripe_customer_id = $1, updated_at = now() WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, stripeCustomerId, userId)
	return err
}
