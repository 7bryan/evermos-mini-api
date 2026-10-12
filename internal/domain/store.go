package domain

import "time"

type Store struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;uniqueIndex"`
	Name      string `gorm:"size:255;not null"`
	PhotoURL  string `gorm:"size:255;not null;default:''"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
