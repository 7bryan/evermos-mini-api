package validator

import (
	"strings"
	"testing"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type registerReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"kata_sandi" validate:"required,min=6,max=72"`
}

func TestStructValid(t *testing.T) {
	if err := New().Struct(registerReq{Email: "a@b.co", Password: "secret1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStructReportsClientFieldNames(t *testing.T) {
	err := New().Struct(registerReq{Email: "not-an-email", Password: "123"})
	appErr, ok := apperrors.As(err)
	if !ok || appErr.Kind != apperrors.KindValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
	joined := strings.Join(appErr.Details, "|")
	if !strings.Contains(joined, "email must be a valid email address") ||
		!strings.Contains(joined, "kata_sandi must be at least 6 characters") {
		t.Fatalf("unexpected details: %v", appErr.Details)
	}
}
