package region

import (
	"context"
	"errors"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Provinces(ctx context.Context) ([]Province, error) {
	v, err := s.provider.Provinces(ctx)
	return v, mapError(err, "province not found")
}

func (s *Service) Province(ctx context.Context, id string) (*Province, error) {
	v, err := s.provider.Province(ctx, id)
	return v, mapError(err, "province not found")
}

func (s *Service) Cities(ctx context.Context, provinceID string) ([]City, error) {
	v, err := s.provider.Cities(ctx, provinceID)
	return v, mapError(err, "province not found")
}

func (s *Service) City(ctx context.Context, id string) (*City, error) {
	v, err := s.provider.City(ctx, id)
	return v, mapError(err, "city not found")
}

func (s *Service) ValidateLocation(ctx context.Context, provinceID, cityID string) error {
	if _, err := s.provider.Province(ctx, provinceID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.Validation("validation failed", "id_provinsi is not a valid province")
		}
		return upstream(err)
	}
	cities, err := s.provider.Cities(ctx, provinceID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.Validation("validation failed", "id_provinsi is not a valid province")
		}
		return upstream(err)
	}
	for _, c := range cities {
		if c.ID == cityID {
			return nil
		}
	}
	return apperrors.Validation("validation failed", "id_kota is not a valid city of the selected province")
}

func mapError(err error, notFoundMessage string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return apperrors.NotFound(notFoundMessage)
	default:
		return upstream(err)
	}
}

func upstream(err error) error {
	return apperrors.Upstream("region service unavailable", err)
}
