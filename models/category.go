package models

// Category represents a product category in the catalog.
// It includes a unique code and display name.
type Category struct {
    ID       uint      `gorm:"primaryKey"`
    Code     string    `gorm:"uniqueIndex;not null"`
    Name     string    `gorm:"not null"`
    Products []Product `gorm:"foreignKey:CategoryID"`
}

func (c *Category) TableName() string {
    return "categories"
}
