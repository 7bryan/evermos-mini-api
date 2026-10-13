package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/7bryan/evermos-mini-api/internal/application/region"
	"github.com/7bryan/evermos-mini-api/internal/domain"
	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

const (
	maxPasswordBytes = 72
	storeNamePrefix  = "Toko "
	birthDateOutput  = "02/01/2006"
	lookupTimeout    = 2 * time.Second
)

var birthDateInputs = []string{"02/01/2006", "2006-01-02"}

type UserRepository interface {
	CreateWithStore(ctx context.Context, user *domain.User, store *domain.Store) error
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type TokenIssuer interface {
	Generate(userID uint) (string, error)
}

type LocationService interface {
	ValidateLocation(ctx context.Context, provinceID, cityID string) error
	Province(ctx context.Context, id string) (*region.Province, error)
	City(ctx context.Context, id string) (*region.City, error)
}

type Service struct {
	users     UserRepository
	hasher    PasswordHasher
	tokens    TokenIssuer
	regions   LocationService
	dummyHash string
	now       func() time.Time
}

func NewService(users UserRepository, hasher PasswordHasher, tokens TokenIssuer, regions LocationService) (*Service, error) {
	dummy, err := hasher.Hash("timing-equalizer-password")
	if err != nil {
		return nil, fmt.Errorf("prepare dummy hash: %w", err)
	}
	return &Service{users: users, hasher: hasher, tokens: tokens, regions: regions, dummyHash: dummy, now: time.Now}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return apperrors.Validation("validation failed", "nama is required")
	}
	if len(req.Password) > maxPasswordBytes {
		return apperrors.Validation("validation failed", "kata_sandi must be at most 72 bytes")
	}
	birth, err := parseBirthDate(req.BirthDate, s.now())
	if err != nil {
		return err
	}
	if err := s.regions.ValidateLocation(ctx, req.ProvinceID, req.CityID); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return apperrors.Internal(err)
	}

	user := &domain.User{
		Name:         name,
		PasswordHash: hash,
		Phone:        strings.TrimSpace(req.Phone),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		BirthDate:    birth,
		Occupation:   strings.TrimSpace(req.Occupation),
		ProvinceID:   req.ProvinceID,
		CityID:       req.CityID,
	}
	store := &domain.Store{Name: storeNamePrefix + name}

	if err := s.users.CreateWithStore(ctx, user, store); err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicatePhone):
			return apperrors.Conflict("phone number is already registered")
		case errors.Is(err, domain.ErrDuplicateEmail):
			return apperrors.Conflict("email is already registered")
		}
		return apperrors.Internal(err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.users.FindByPhone(ctx, strings.TrimSpace(req.Phone))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.hasher.Compare(s.dummyHash, req.Password)
			return nil, invalidCredentials()
		}
		return nil, apperrors.Internal(err)
	}
	if !s.hasher.Compare(user.PasswordHash, req.Password) {
		return nil, invalidCredentials()
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	province, city := s.locationRefs(ctx, user)
	return &LoginResponse{
		Name:       user.Name,
		Phone:      user.Phone,
		BirthDate:  user.BirthDate.Format(birthDateOutput),
		Occupation: user.Occupation,
		Email:      user.Email,
		Province:   province,
		City:       city,
		Token:      token,
	}, nil
}

func (s *Service) locationRefs(ctx context.Context, u *domain.User) (LocationRef, LocationRef) {
	ctx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()

	province := LocationRef{ID: u.ProvinceID}
	city := LocationRef{ID: u.CityID}
	if p, err := s.regions.Province(ctx, u.ProvinceID); err == nil {
		province.Name = p.Name
	} else {
		log.Printf("login: province lookup %s: %v", u.ProvinceID, err)
	}
	if c, err := s.regions.City(ctx, u.CityID); err == nil {
		city.Name = c.Name
	} else {
		log.Printf("login: city lookup %s: %v", u.CityID, err)
	}
	return province, city
}

func invalidCredentials() error {
	return apperrors.Unauthorized("invalid phone number or password")
}

func parseBirthDate(raw string, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range birthDateInputs {
		t, err := time.Parse(layout, raw)
		if err != nil {
			continue
		}
		if t.After(now) || t.Year() < 1900 {
			break
		}
		return t, nil
	}
	return time.Time{}, apperrors.Validation("validation failed",
		"tanggal_Lahir must be a valid past date (DD/MM/YYYY or YYYY-MM-DD)")
}
