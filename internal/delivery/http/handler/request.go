package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/delivery/http/validator"
	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

func bind(c *fiber.Ctx, val *validator.Validator, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return apperrors.Validation("invalid request body")
	}
	return val.Struct(dst)
}
