package domain

import "time"

type Category struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
