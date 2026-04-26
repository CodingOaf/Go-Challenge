package models

import (
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) GetProducts(offset, limit int, category string, priceLessThan *float64) ([]Product, int64, error) {
	var products []Product
	query := r.db.Model(&Product{}).Preload("Category").Preload("Variants")
	countQuery := r.db.Model(&Product{})

	if category != "" {
		query = query.Joins("Category").Where("categories.code = ?", category)
		countQuery = countQuery.Joins("Category").Where("categories.code = ?", category)
	}

	if priceLessThan != nil {
		query = query.Where("products.price < ?", *priceLessThan)
		countQuery = countQuery.Where("products.price < ?", *priceLessThan)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("products.id asc").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").Where("code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
