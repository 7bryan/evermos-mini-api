package region

import (
	"context"
	"errors"
	"strings"
	"testing"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type fakeProvider struct {
	err error
}

var (
	testProvinces = []Province{{ID: "11", Name: "ACEH"}, {ID: "12", Name: "SUMATERA UTARA"}}
	testCities    = map[string][]City{
		"11": {{ID: "1101", ProvinceID: "11", Name: "SIMEULUE"}},
		"12": {{ID: "1201", ProvinceID: "12", Name: "NIAS"}},
	}
)

func (f fakeProvider) Provinces(context.Context) ([]Province, error) {
	return testProvinces, f.err
}

func (f fakeProvider) Province(_ context.Context, id string) (*Province, error) {
	if f.err != nil {
		return nil, f.err
	}
	for _, p := range testProvinces {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, ErrNotFound
}

func (f fakeProvider) Cities(_ context.Context, provinceID string) ([]City, error) {
	if f.err != nil {
		return nil, f.err
	}
	list, ok := testCities[provinceID]
	if !ok {
		return nil, ErrNotFound
	}
	return list, nil
}

func (f fakeProvider) City(_ context.Context, id string) (*City, error) {
	if f.err != nil {
		return nil, f.err
	}
	for _, list := range testCities {
		for _, c := range list {
			if c.ID == id {
				return &c, nil
			}
		}
	}
	return nil, ErrNotFound
}

func kindOf(t *testing.T, err error) (apperrors.Kind, []string) {
	t.Helper()
	appErr, ok := apperrors.As(err)
	if !ok {
		t.Fatalf("expected *apperrors.Error, got %v", err)
	}
	return appErr.Kind, appErr.Details
}

func TestValidateLocation(t *testing.T) {
	cases := []struct {
		name         string
		provider     fakeProvider
		provinceID   string
		cityID       string
		wantKind     apperrors.Kind
		wantOK       bool
		wantContains string
	}{
		{name: "valid", provinceID: "11", cityID: "1101", wantOK: true},
		{name: "unknown province", provinceID: "99", cityID: "1101",
			wantKind: apperrors.KindValidation, wantContains: "id_provinsi"},
		{name: "unknown city", provinceID: "11", cityID: "9999",
			wantKind: apperrors.KindValidation, wantContains: "id_kota"},
		{name: "city of another province", provinceID: "11", cityID: "1201",
			wantKind: apperrors.KindValidation, wantContains: "id_kota"},
		{name: "upstream down", provider: fakeProvider{err: errors.New("timeout")}, provinceID: "11", cityID: "1101",
			wantKind: apperrors.KindUpstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewService(tc.provider).ValidateLocation(context.Background(), tc.provinceID, tc.cityID)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			kind, details := kindOf(t, err)
			if kind != tc.wantKind {
				t.Fatalf("kind = %v, want %v", kind, tc.wantKind)
			}
			if tc.wantContains != "" && !strings.Contains(strings.Join(details, "|"), tc.wantContains) {
				t.Fatalf("details = %v, want mention of %s", details, tc.wantContains)
			}
		})
	}
}

func TestServiceMapsNotFoundAndUpstream(t *testing.T) {
	svc := NewService(fakeProvider{})
	if _, err := svc.City(context.Background(), "9999"); true {
		if kind, _ := kindOf(t, err); kind != apperrors.KindNotFound {
			t.Fatalf("kind = %v, want NotFound", kind)
		}
	}
	if _, err := svc.Cities(context.Background(), "99"); true {
		if kind, _ := kindOf(t, err); kind != apperrors.KindNotFound {
			t.Fatalf("kind = %v, want NotFound", kind)
		}
	}

	down := NewService(fakeProvider{err: errors.New("timeout")})
	if _, err := down.Provinces(context.Background()); true {
		if kind, _ := kindOf(t, err); kind != apperrors.KindUpstream {
			t.Fatalf("kind = %v, want Upstream", kind)
		}
	}
}
