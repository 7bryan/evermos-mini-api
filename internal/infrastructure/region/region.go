package region

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("region not found")

type Province struct {
	ID   string
	Name string
}

type City struct {
	ID         string
	ProvinceID string
	Name       string
}

type Provider interface {
	Provinces(ctx context.Context) ([]Province, error)
	Province(ctx context.Context, id string) (*Province, error)
	Cities(ctx context.Context, provinceID string) ([]City, error)
	City(ctx context.Context, id string) (*City, error)
}
