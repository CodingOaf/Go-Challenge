package models

// Product represents a product in the catalog.
// It includes a unique code, price, category, and variants.
type Product struct {
	ID         uint      `gorm:"primaryKey"`
	Code       string    `gorm:"uniqueIndex;not null"`
	Price      float64   `gorm:"type:decimal(10,2);not null"`
	CategoryID uint      `gorm:"not null"`
	Category   Category  `gorm:"foreignKey:CategoryID"`
	Variants   []Variant `gorm:"foreignKey:ProductID"`
}

func (p *Product) TableName() string {
	return "products"
}
