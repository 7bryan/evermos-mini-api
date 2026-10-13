package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/7bryan/evermos-mini-api/internal/application/region"
	"github.com/7bryan/evermos-mini-api/internal/domain"
	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type fakeRepo struct {
	users     map[string]*domain.User
	stores    []*domain.Store
	createErr error
	created   int
}

func (f *fakeRepo) CreateWithStore(_ context.Context, u *domain.User, s *domain.Store) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created++
	u.ID = uint(f.created)
	s.UserID = u.ID
	f.users[u.Phone] = u
	f.stores = append(f.stores, s)
	return nil
}

func (f *fakeRepo) FindByPhone(_ context.Context, phone string) (*domain.User, error) {
	if u, ok := f.users[phone]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}

type fakeHasher struct{ compares int }

func (h *fakeHasher) Hash(p string) (string, error) { return "hashed:" + p, nil }
func (h *fakeHasher) Compare(hash, p string) bool {
	h.compares++
	return hash == "hashed:"+p
}

type fakeTokens struct{}

func (fakeTokens) Generate(id uint) (string, error) { return fmt.Sprintf("token-%d", id), nil }

type fakeLocator struct{ validateErr, lookupErr error }

func (f fakeLocator) ValidateLocation(context.Context, string, string) error { return f.validateErr }
func (f fakeLocator) Province(_ context.Context, id string) (*region.Province, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return &region.Province{ID: id, Name: "ACEH"}, nil
}
func (f fakeLocator) City(_ context.Context, id string) (*region.City, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return &region.City{ID: id, Name: "KABUPATEN SIMEULUE"}, nil
}

func newService(t *testing.T, loc fakeLocator) (*Service, *fakeRepo, *fakeHasher) {
	t.Helper()
	repo := &fakeRepo{users: map[string]*domain.User{}}
	hasher := &fakeHasher{}
	svc, err := NewService(repo, hasher, fakeTokens{}, loc)
	if err != nil {
		t.Fatal(err)
	}
	svc.now = func() time.Time { return time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC) }
	return svc, repo, hasher
}

func validReq() RegisterRequest {
	return RegisterRequest{
		Name: "Budi", Password: "secret123", Phone: "081234567890", BirthDate: "17/08/1990",
		Occupation: "Engineer", Email: "Budi@Example.com", ProvinceID: "11", CityID: "1101",
	}
}

func kindOf(t *testing.T, err error) apperrors.Kind {
	t.Helper()
	appErr, ok := apperrors.As(err)
	if !ok {
		t.Fatalf("expected *apperrors.Error, got %v", err)
	}
	return appErr.Kind
}

func TestRegisterCreatesUserAndStore(t *testing.T) {
	svc, repo, _ := newService(t, fakeLocator{})
	if err := svc.Register(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	u := repo.users["081234567890"]
	if u == nil {
		t.Fatal("user not stored")
	}
	if u.PasswordHash != "hashed:secret123" {
		t.Fatalf("password not hashed: %q", u.PasswordHash)
	}
	if u.Email != "budi@example.com" {
		t.Fatalf("email not normalized: %q", u.Email)
	}
	if u.IsAdmin {
		t.Fatal("registered user must not be admin")
	}
	if !u.BirthDate.Equal(time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("birth date = %v", u.BirthDate)
	}
	if len(repo.stores) != 1 || repo.stores[0].Name != "Toko Budi" || repo.stores[0].UserID != u.ID {
		t.Fatalf("store not created correctly: %+v", repo.stores)
	}
}

func TestRegisterBirthDates(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"17/08/1990", true},
		{"1990-08-17", true},
		{"31/02/1990", false},
		{"2999-01-01", false},
		{"13/10/2026", false},
		{"01/01/1850", false},
		{"garbage", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			svc, _, _ := newService(t, fakeLocator{})
			req := validReq()
			req.BirthDate = tc.in
			err := svc.Register(context.Background(), req)
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok && kindOf(t, err) != apperrors.KindValidation {
				t.Fatalf("err = %v, want validation", err)
			}
		})
	}
}

func TestRegisterRejectsBadLocationWithoutInserting(t *testing.T) {
	bad := apperrors.Validation("validation failed", "id_kota is not a valid city of the selected province")
	svc, repo, _ := newService(t, fakeLocator{validateErr: bad})
	err := svc.Register(context.Background(), validReq())
	if kindOf(t, err) != apperrors.KindValidation {
		t.Fatalf("err = %v", err)
	}
	if repo.created != 0 {
		t.Fatal("user inserted despite invalid location")
	}
}

func TestRegisterPropagatesUpstreamFailure(t *testing.T) {
	svc, _, _ := newService(t, fakeLocator{validateErr: apperrors.Upstream("region service unavailable", errors.New("x"))})
	if kindOf(t, svc.Register(context.Background(), validReq())) != apperrors.KindUpstream {
		t.Fatal("expected upstream error")
	}
}

func TestRegisterDuplicates(t *testing.T) {
	for name, dupErr := range map[string]error{
		"phone": domain.ErrDuplicatePhone,
		"email": domain.ErrDuplicateEmail,
	} {
		t.Run(name, func(t *testing.T) {
			svc, repo, _ := newService(t, fakeLocator{})
			repo.createErr = dupErr
			err := svc.Register(context.Background(), validReq())
			if kindOf(t, err) != apperrors.KindConflict {
				t.Fatalf("err = %v, want conflict", err)
			}
		})
	}
}

func TestRegisterUnexpectedRepoErrorIsInternal(t *testing.T) {
	svc, repo, _ := newService(t, fakeLocator{})
	repo.createErr = errors.New("Error 1213: deadlock")
	appErr, _ := apperrors.As(svc.Register(context.Background(), validReq()))
	if appErr.Kind != apperrors.KindInternal || appErr.Message != "internal server error" {
		t.Fatalf("err = %+v", appErr)
	}
}

func TestRegisterRejectsOverlongPasswordBytes(t *testing.T) {
	svc, _, _ := newService(t, fakeLocator{})
	req := validReq()
	req.Password = string(make([]byte, 73))
	if kindOf(t, svc.Register(context.Background(), req)) != apperrors.KindValidation {
		t.Fatal("73-byte password accepted")
	}
}

func registered(t *testing.T) (*Service, *fakeHasher) {
	t.Helper()
	svc, _, hasher := newService(t, fakeLocator{})
	if err := svc.Register(context.Background(), validReq()); err != nil {
		t.Fatal(err)
	}
	return svc, hasher
}

func TestLoginSuccess(t *testing.T) {
	svc, _ := registered(t)
	res, err := svc.Login(context.Background(), LoginRequest{Phone: "081234567890", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Token != "token-1" || res.Province.Name != "ACEH" || res.City.Name != "KABUPATEN SIMEULUE" ||
		res.BirthDate != "17/08/1990" || res.Email != "budi@example.com" {
		t.Fatalf("response = %+v", res)
	}
}

func TestLoginFailuresLookIdentical(t *testing.T) {
	svc, hasher := registered(t)

	_, wrongPass := svc.Login(context.Background(), LoginRequest{Phone: "081234567890", Password: "nope"})
	hasher.compares = 0
	_, unknown := svc.Login(context.Background(), LoginRequest{Phone: "089999999999", Password: "secret123"})

	a, _ := apperrors.As(wrongPass)
	b, _ := apperrors.As(unknown)
	if a == nil || b == nil || a.Kind != apperrors.KindUnauthorized || b.Kind != apperrors.KindUnauthorized {
		t.Fatalf("wrongPass=%v unknown=%v", wrongPass, unknown)
	}
	if a.Message != b.Message {
		t.Fatalf("messages differ: %q vs %q", a.Message, b.Message)
	}
	if hasher.compares != 1 {
		t.Fatalf("unknown phone did %d hash comparisons, want 1 (timing equalizer)", hasher.compares)
	}
}

func TestLoginSurvivesRegionOutage(t *testing.T) {
	svc, repo, _ := newService(t, fakeLocator{lookupErr: errors.New("emsifa down")})
	_ = svc.Register(context.Background(), validReq())
	res, err := svc.Login(context.Background(), LoginRequest{Phone: "081234567890", Password: "secret123"})
	if err != nil {
		t.Fatalf("login failed during region outage: %v", err)
	}
	if res.Province.ID != "11" || res.Province.Name != "" || res.Token == "" || repo.created != 1 {
		t.Fatalf("response = %+v", res)
	}
}
