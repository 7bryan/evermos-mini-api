package apperrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestAsFindsWrappedError(t *testing.T) {
	wrapped := fmt.Errorf("service: %w", NotFound("store not found"))
	got, ok := As(wrapped)
	if !ok || got.Kind != KindNotFound {
		t.Fatalf("expected wrapped NotFound, got %v (ok=%v)", got, ok)
	}
}

func TestInternalHidesCause(t *testing.T) {
	cause := errors.New("Error 1062: Duplicate entry")
	e := Internal(cause)
	if e.Message != "internal server error" {
		t.Fatalf("message leaks details: %q", e.Message)
	}
	if !errors.Is(e, cause) {
		t.Fatal("cause should be reachable via Unwrap for logging")
	}
}
