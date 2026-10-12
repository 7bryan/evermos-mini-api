package domain

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Phone        string    `gorm:"size:20;not null;uniqueIndex"`
	Email        string    `gorm:"size:255;not null;uniqueIndex"`
	BirthDate    time.Time `gorm:"type:date;not null"`
	Occupation   string    `gorm:"size:100;not null"`
	ProvinceID   string    `gorm:"size:20;not null"`
	CityID       string    `gorm:"size:20;not null"`
	IsAdmin      bool      `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Store *Store `gorm:"constraint:OnDelete:CASCADE"`
}
