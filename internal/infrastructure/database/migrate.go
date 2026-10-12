package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/7bryan/evermos-mini-api/internal/domain"
)

func Models() []any {
	return []any{
		&domain.User{},
		&domain.Store{},
		&domain.Address{},
		&domain.Category{},
		&domain.Product{},
		&domain.ProductPhoto{},
		&domain.ProductLog{},
		&domain.ProductLogPhoto{},
		&domain.Transaction{},
		&domain.TransactionItem{},
	}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(Models()...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
