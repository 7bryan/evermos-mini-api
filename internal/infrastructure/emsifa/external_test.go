//go:build external

package emsifa

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRealEmsifaAPI(t *testing.T) {
	base := os.Getenv("REGION_API_BASE_URL")
	if base == "" {
		base = DefaultBaseURL
	}
	c := NewClient(base, 10*time.Second, time.Hour)
	ctx := context.Background()

	provinces, err := c.Provinces(ctx)
	if err != nil || len(provinces) < 30 {
		t.Fatalf("provinces: n=%d err=%v", len(provinces), err)
	}
	t.Logf("first province: %+v", provinces[0])

	cities, err := c.Cities(ctx, provinces[0].ID)
	if err != nil || len(cities) == 0 {
		t.Fatalf("cities: n=%d err=%v", len(cities), err)
	}
	t.Logf("first city: %+v", cities[0])
	if cities[0].ProvinceID != provinces[0].ID {
		t.Fatalf("city %+v does not belong to province %s", cities[0], provinces[0].ID)
	}

	city, err := c.City(ctx, cities[0].ID)
	if err != nil || city.ID != cities[0].ID {
		t.Fatalf("city detail: %+v err=%v", city, err)
	}
}
