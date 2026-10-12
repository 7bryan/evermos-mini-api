package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestGenerateAndParse(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)
	token, err := m.Generate(42)
	if err != nil {
		t.Fatal(err)
	}
	id, err := m.Parse(token)
	if err != nil || id != 42 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestParseRejectsExpired(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)
	m.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	token, _ := m.Generate(1)
	m.now = time.Now
	if _, err := m.Parse(token); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	token, _ := NewTokenManager(testSecret, time.Hour).Generate(1)
	other := NewTokenManager("ffffffffffffffffffffffffffffffff", time.Hour)
	if _, err := other.Parse(token); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsAlgNone(t *testing.T) {
	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTokenManager(testSecret, time.Hour).Parse(token); err != ErrInvalidToken {
		t.Fatalf("alg=none token accepted: %v", err)
	}
}

func TestParseRejectsMissingExpiry(t *testing.T) {
	claims := jwt.RegisteredClaims{Issuer: tokenIssuer, Subject: "1"}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if _, err := NewTokenManager(testSecret, time.Hour).Parse(token); err != ErrInvalidToken {
		t.Fatalf("token without exp accepted: %v", err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)
	for _, s := range []string{"", "abc", "a.b.c"} {
		if _, err := m.Parse(s); err != ErrInvalidToken {
			t.Fatalf("%q accepted", s)
		}
	}
}
