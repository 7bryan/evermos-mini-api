package domain

import "time"

type ProductLog struct {
	ID            uint   `gorm:"primaryKey"`
	ProductID     uint   `gorm:"not null;index"`
	StoreID       uint   `gorm:"not null;index"`
	CategoryID    uint   `gorm:"not null;index"`
	Name          string `gorm:"size:255;not null"`
	Slug          string `gorm:"size:300;not null"`
	ResellerPrice uint   `gorm:"not null"`
	ConsumerPrice uint   `gorm:"not null"`
	Description   string `gorm:"type:text;not null"`
	CreatedAt     time.Time

	Store    Store             `gorm:"constraint:OnDelete:RESTRICT"`
	Category Category          `gorm:"constraint:OnDelete:RESTRICT"`
	Photos   []ProductLogPhoto `gorm:"constraint:OnDelete:CASCADE"`
}

type ProductLogPhoto struct {
	ID           uint   `gorm:"primaryKey"`
	ProductLogID uint   `gorm:"not null;index"`
	URL          string `gorm:"size:255;not null"`
}

type Transaction struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `gorm:"not null;index"`
	AddressID     uint   `gorm:"not null;index"`
	InvoiceCode   string `gorm:"size:50;not null;uniqueIndex"`
	TotalPrice    uint   `gorm:"not null"`
	PaymentMethod string `gorm:"size:50;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	User    User              `gorm:"constraint:OnDelete:RESTRICT" json:"-"`
	Address Address           `gorm:"constraint:OnDelete:RESTRICT"`
	Items   []TransactionItem `gorm:"constraint:OnDelete:CASCADE"`
}

type TransactionItem struct {
	ID            uint `gorm:"primaryKey"`
	TransactionID uint `gorm:"not null;index"`
	ProductLogID  uint `gorm:"not null;index"`
	Quantity      uint `gorm:"not null;check:chk_transaction_items_quantity,quantity > 0"`
	TotalPrice    uint `gorm:"not null"`

	ProductLog ProductLog `gorm:"constraint:OnDelete:RESTRICT"`
}
