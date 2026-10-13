package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/delivery/http/response"
	"github.com/7bryan/evermos-mini-api/internal/domain"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/security"
)

type fakeChecker map[uint]bool

func (f fakeChecker) IsAdmin(_ context.Context, id uint) (bool, error) {
	v, ok := f[id]
	if !ok {
		return false, domain.ErrNotFound
	}
	return v, nil
}

type brokenChecker struct{}

func (brokenChecker) IsAdmin(context.Context, uint) (bool, error) {
	return false, errors.New("db down")
}

func newApp(tm *security.TokenManager, checker AdminChecker) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Get("/me", Authenticate(tm), func(c *fiber.Ctx) error {
		id, _ := UserID(c)
		return c.SendString(strconv.Itoa(int(id)))
	})
	app.Get("/admin", Authenticate(tm), RequireAdmin(checker), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/admin-without-auth", RequireAdmin(checker), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func status(t *testing.T, app *fiber.App, path string, headers map[string]string) int {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestAuthenticate(t *testing.T) {
	tm := security.NewTokenManager("0123456789abcdef0123456789abcdef", time.Hour)
	good, _ := tm.Generate(7)
	app := newApp(tm, fakeChecker{})

	cases := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{"raw token header", map[string]string{"token": good}, 200},
		{"bearer in token header", map[string]string{"token": "Bearer " + good}, 200},
		{"authorization bearer", map[string]string{"Authorization": "Bearer " + good}, 200},
		{"no credentials", nil, 401},
		{"garbage token", map[string]string{"token": "abc.def.ghi"}, 401},
		{"empty bearer", map[string]string{"Authorization": "Bearer "}, 401},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := status(t, app, "/me", tc.headers); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestUserIDComesFromToken(t *testing.T) {
	tm := security.NewTokenManager("0123456789abcdef0123456789abcdef", time.Hour)
	tok, _ := tm.Generate(7)
	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("token", tok)
	resp, _ := newApp(tm, fakeChecker{}).Test(req)
	buf := make([]byte, 8)
	n, _ := resp.Body.Read(buf)
	if string(buf[:n]) != "7" {
		t.Fatalf("user id = %q, want 7", buf[:n])
	}
}

func TestRequireAdmin(t *testing.T) {
	tm := security.NewTokenManager("0123456789abcdef0123456789abcdef", time.Hour)
	admin, _ := tm.Generate(1)
	regular, _ := tm.Generate(2)
	deleted, _ := tm.Generate(3)
	app := newApp(tm, fakeChecker{1: true, 2: false})

	if got := status(t, app, "/admin", map[string]string{"token": admin}); got != 200 {
		t.Fatalf("admin: %d", got)
	}
	if got := status(t, app, "/admin", map[string]string{"token": regular}); got != 403 {
		t.Fatalf("regular user: %d, want 403", got)
	}
	if got := status(t, app, "/admin", map[string]string{"token": deleted}); got != 401 {
		t.Fatalf("deleted user: %d, want 401", got)
	}
	if got := status(t, app, "/admin", nil); got != 401 {
		t.Fatalf("anonymous: %d, want 401", got)
	}
	if got := status(t, app, "/admin-without-auth", nil); got != 401 {
		t.Fatalf("RequireAdmin without Authenticate: %d, want 401", got)
	}

	broken := newApp(tm, brokenChecker{})
	if got := status(t, broken, "/admin", map[string]string{"token": admin}); got != 500 {
		t.Fatalf("db failure: %d, want 500", got)
	}
}
