package models

import "gorm.io/gorm"

type CategoriesRepository struct {
    db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
    return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) GetAllCategories() ([]Category, error) {
    var categories []Category
    if err := r.db.Order("id asc").Find(&categories).Error; err != nil {
        return nil, err
    }
    return categories, nil
}

func (r *CategoriesRepository) CreateCategory(category *Category) error {
    return r.db.Create(category).Error
}
