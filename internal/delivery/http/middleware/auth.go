package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/domain"
	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

const localsUserID = "userID"

type TokenParser interface {
	Parse(token string) (uint, error)
}

type AdminChecker interface {
	IsAdmin(ctx context.Context, userID uint) (bool, error)
}

func Authenticate(parser TokenParser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := stripBearer(c.Get("token"))
		if raw == "" {
			raw = stripBearer(c.Get(fiber.HeaderAuthorization))
		}
		if raw == "" {
			return apperrors.Unauthorized("missing token")
		}
		id, err := parser.Parse(raw)
		if err != nil {
			return apperrors.Unauthorized("invalid or expired token")
		}
		c.Locals(localsUserID, id)
		return c.Next()
	}
}

func RequireAdmin(checker AdminChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := UserID(c)
		if !ok {
			return apperrors.Unauthorized("authentication required")
		}
		isAdmin, err := checker.IsAdmin(c.UserContext(), id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return apperrors.Unauthorized("invalid or expired token")
			}
			return apperrors.Internal(err)
		}
		if !isAdmin {
			return apperrors.Forbidden("admin access required")
		}
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) (uint, bool) {
	id, ok := c.Locals(localsUserID).(uint)
	return id, ok && id != 0
}

func stripBearer(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 7 && strings.EqualFold(v[:7], "bearer ") {
		v = strings.TrimSpace(v[7:])
	}
	return v
}
