package security

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "evermos-mini-api"

var ErrInvalidToken = errors.New("invalid or expired token")

type TokenManager struct {
	secret []byte
	expiry time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, expiry time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), expiry: expiry, now: time.Now}
}

func (m *TokenManager) Generate(userID uint) (string, error) {
	now := m.now()
	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) Parse(tokenString string) (uint, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return 0, ErrInvalidToken
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil || id == 0 {
		return 0, ErrInvalidToken
	}
	return uint(id), nil
}
