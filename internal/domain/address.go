package domain

import "time"

type Address struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null;index"`
	Title        string `gorm:"size:100;not null"`
	ReceiverName string `gorm:"size:255;not null"`
	Phone        string `gorm:"size:20;not null"`
	Detail       string `gorm:"type:text;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	User *User `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}
