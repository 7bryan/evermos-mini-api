package domain

import "time"

type Product struct {
	ID            uint   `gorm:"primaryKey"`
	StoreID       uint   `gorm:"not null;index"`
	CategoryID    uint   `gorm:"not null;index"`
	Name          string `gorm:"size:255;not null;index"`
	Slug          string `gorm:"size:300;not null;uniqueIndex"`
	ResellerPrice uint   `gorm:"not null"`
	ConsumerPrice uint   `gorm:"not null"`
	Stock         uint   `gorm:"not null"`
	Description   string `gorm:"type:text;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Store    Store          `gorm:"constraint:OnDelete:CASCADE"`
	Category Category       `gorm:"constraint:OnDelete:RESTRICT"`
	Photos   []ProductPhoto `gorm:"constraint:OnDelete:CASCADE"`
}

type ProductPhoto struct {
	ID        uint   `gorm:"primaryKey"`
	ProductID uint   `gorm:"not null;index"`
	URL       string `gorm:"size:255;not null"`
	CreatedAt time.Time
}
