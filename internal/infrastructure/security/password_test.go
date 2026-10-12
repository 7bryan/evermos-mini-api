package security

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCompare(t *testing.T) {
	h := NewPasswordHasher(bcrypt.MinCost)
	hash, err := h.Hash("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse" {
		t.Fatal("password stored in plaintext")
	}
	if !h.Compare(hash, "correct horse") {
		t.Fatal("correct password rejected")
	}
	if h.Compare(hash, "wrong") {
		t.Fatal("wrong password accepted")
	}
}

func TestHashRejectsOver72Bytes(t *testing.T) {
	h := NewPasswordHasher(bcrypt.MinCost)
	if _, err := h.Hash(strings.Repeat("a", 73)); err != ErrPasswordTooLong {
		t.Fatalf("err = %v, want ErrPasswordTooLong", err)
	}
}

func TestCompareWithGarbageHash(t *testing.T) {
	if NewPasswordHasher(bcrypt.MinCost).Compare("not-a-hash", "x") {
		t.Fatal("garbage hash must not match")
	}
}
