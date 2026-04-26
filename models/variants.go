package models

// Variant represents a product variant in the catalog.
// It includes a unique name, SKU, and an optional price.
// Variants without a price inherit the parent product's price.
type Variant struct {
	ID        uint     `gorm:"primaryKey"`
	ProductID uint     `gorm:"not null"`
	Name      string   `gorm:"not null"`
	SKU       string   `gorm:"uniqueIndex;not null"`
	Price     *float64 `gorm:"type:decimal(10,2);null"`
}

func (v *Variant) TableName() string {
	return "product_variants"
}
