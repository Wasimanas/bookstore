package product

import (
	"context"
	"database/sql"
	"errors"
	"store/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, Product models.Product) (models.Product, error)
	GetByID(ctx context.Context, id int64) (models.Product, error)
	Update(ctx context.Context, id int64, Product models.Product) (models.Product, error)
	Delete(ctx context.Context, id int64) error
	Search(ctx context.Context, title string) ([]models.Product, error)
    GetAllProducts(ctx context.Context, limit int, offset int) ([]models.Product, error)
}

type PostgresProductRepo struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) ProductRepository {
	return &PostgresProductRepo{db: db}
}

func (r *PostgresProductRepo) GetAllProducts(ctx context.Context, limit int, offset int)  ([]models.Product, error) { 
	query := `SELECT id, title, author, year, price, created_at, updated_at FROM Products order by updated_at desc limit $1 offset $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Products []models.Product
	for rows.Next() {
		var b models.Product
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Price, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		Products = append(Products, b)
	}

	return Products, nil


}


func (r *PostgresProductRepo) Create(ctx context.Context, Product models.Product) (models.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Product, err
	}

	query := `INSERT INTO Products (title, author, year, price) VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRowContext(ctx, query, Product.Title, Product.Author, Product.Year, Product.Price).Scan(&Product.ID)
	if err != nil {
		tx.Rollback()
		return Product, err
	}

	if err := tx.Commit(); err != nil {
		return Product, err
	}

	return Product, nil
}

func (r *PostgresProductRepo) GetByID(ctx context.Context, id int64) (models.Product, error) {
	var Product models.Product
	query := `SELECT id, title, author, year, price FROM Products WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&Product.ID, &Product.Title, &Product.Author, &Product.Year, &Product.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Product, nil
		}
		return Product, err
	}
	return Product, nil
}

func (r *PostgresProductRepo) Update(ctx context.Context, id int64, Product models.Product) (models.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Product, err
	}

	query := `UPDATE Products SET title=$1, author=$2, year=$3, price=$4 WHERE id=$5`
	res, err := tx.ExecContext(ctx, query, Product.Title, Product.Author, Product.Year, Product.Price, id)
	if err != nil {
		tx.Rollback()
		return Product, err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return Product, errors.New("no Product found to update")
	}

	if err := tx.Commit(); err != nil {
		return Product, err
	}

	Product.ID = id
	return Product, nil
}

func (r *PostgresProductRepo) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := `DELETE FROM Products WHERE id=$1`
	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return errors.New("no Product found to delete")
	}

	return tx.Commit()
}

func (r *PostgresProductRepo) Search(ctx context.Context, title string) ([]models.Product, error) {
	query := `SELECT id, title, author, year, price FROM Products WHERE title ILIKE '%' || $1 || '%'`
	rows, err := r.db.QueryContext(ctx, query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Products []models.Product
	for rows.Next() {
		var b models.Product
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Price); err != nil {
			return nil, err
		}
		Products = append(Products, b)
	}

	return Products, nil
}

