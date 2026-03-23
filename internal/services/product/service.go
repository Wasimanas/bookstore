package product

import (
	"context"
	"errors"
	"store/internal/models"
	"store/internal/repositories/product"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	GetProductByID(ctx context.Context, id int64) (models.Product, error)
	UpdateProduct(ctx context.Context, id int64, product models.Product) (models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	SearchProducts(ctx context.Context, title string) ([]models.Product, error)
	GetAllProducts(ctx context.Context, limit int, offset int) ([]models.Product, error)
}

type ProductServiceImpl struct {
	repo product.ProductRepository
}

func NewProductsService(repo product.ProductRepository) ProductService {
	return &ProductServiceImpl{repo: repo}
}

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	if product.Title == "" || product.Author == "" {
		return models.Product{}, errors.New("title and author cannot be empty")
	}
	if product.Price < 0 {
		return models.Product{}, errors.New("price cannot be negative")
	}

	return s.repo.Create(ctx, product)
}

func (s *ProductServiceImpl) GetProductByID(ctx context.Context, id int64) (models.Product, error) {
	if id <= 0 {
		return models.Product{}, errors.New("invalid product ID")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *ProductServiceImpl) UpdateProduct(ctx context.Context, id int64, product models.Product) (models.Product, error) {
	if id <= 0 {
		return models.Product{}, errors.New("invalid product ID")
	}
	if product.Title == "" || product.Author == "" {
		return models.Product{}, errors.New("title and author cannot be empty")
	}
	if product.Price < 0 {
		return models.Product{}, errors.New("price cannot be negative")
	}

	return s.repo.Update(ctx, id, product)
}

func (s *ProductServiceImpl) DeleteProduct(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid product ID")
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProductServiceImpl) SearchProducts(ctx context.Context, title string) ([]models.Product, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	return s.repo.Search(ctx, title)
}

func (s *ProductServiceImpl) GetAllProducts(ctx context.Context, limit int, offset int) ([]models.Product, error) {
	if limit == 0 {
		limit = 10
	}
	return s.repo.GetAllProducts(ctx, limit, offset)
}

