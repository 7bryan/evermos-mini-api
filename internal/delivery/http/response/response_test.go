package response

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

func TestErrorHandlerMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"validation", apperrors.Validation("validation failed", "email is required"), 400, "validation failed"},
		{"unauthorized", apperrors.Unauthorized("invalid token"), 401, "invalid token"},
		{"forbidden", apperrors.Forbidden("admin only"), 403, "admin only"},
		{"not found", apperrors.NotFound("no such store"), 404, "no such store"},
		{"conflict", apperrors.Conflict("email already used"), 409, "email already used"},
		{"upstream", apperrors.Upstream("region service unavailable", errors.New("timeout")), 502, "region service unavailable"},
		{"internal hides cause", apperrors.Internal(errors.New("Error 1062 secret")), 500, "internal server error"},
		{"unknown error", errors.New("boom"), 500, "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
			app.Get("/x", func(*fiber.Ctx) error { return tc.err })

			resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			var env Envelope
			if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if env.Status || env.Message != tc.wantMsg {
				t.Fatalf("envelope = %+v", env)
			}
			if strings.Contains(env.Message, "1062") {
				t.Fatal("internal details leaked")
			}
		})
	}
}

func TestUnknownRouteUsesEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	resp, _ := app.Test(httptest.NewRequest("GET", "/nope", nil))
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var env Envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil || env.Status {
		t.Fatalf("expected failure envelope, err=%v env=%+v", err, env)
	}
}
