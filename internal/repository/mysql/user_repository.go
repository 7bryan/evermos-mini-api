package mysql

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/7bryan/evermos-mini-api/internal/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateWithStore(ctx context.Context, user *domain.User, store *domain.Store) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		store.UserID = user.ID
		return tx.Create(store).Error
	})
	return translateUserError(err)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) IsAdmin(ctx context.Context, userID uint) (bool, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Select("is_admin").First(&u, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, domain.ErrNotFound
	}
	if err != nil {
		return false, err
	}
	return u.IsAdmin, nil
}

func translateUserError(err error) error {
	key, dup := duplicateKey(err)
	if !dup {
		return err
	}
	switch {
	case strings.Contains(key, "phone"):
		return domain.ErrDuplicatePhone
	case strings.Contains(key, "email"):
		return domain.ErrDuplicateEmail
	}
	return err
}
