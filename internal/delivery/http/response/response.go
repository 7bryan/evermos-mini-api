package response

import (
	"errors"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type Envelope struct {
	Status  bool     `json:"status"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
	Data    any      `json:"data"`
}

func Success(c *fiber.Ctx, code int, message string, data any) error {
	return c.Status(code).JSON(Envelope{Status: true, Message: message, Data: data})
}

func OK(c *fiber.Ctx, message string, data any) error {
	return Success(c, http.StatusOK, message, data)
}

func Created(c *fiber.Ctx, message string, data any) error {
	return Success(c, http.StatusCreated, message, data)
}

func fail(c *fiber.Ctx, code int, message string, details []string) error {
	return c.Status(code).JSON(Envelope{Status: false, Message: message, Errors: details})
}

func statusFor(kind apperrors.Kind) int {
	switch kind {
	case apperrors.KindValidation:
		return http.StatusBadRequest
	case apperrors.KindUnauthorized:
		return http.StatusUnauthorized
	case apperrors.KindForbidden:
		return http.StatusForbidden
	case apperrors.KindNotFound:
		return http.StatusNotFound
	case apperrors.KindConflict:
		return http.StatusConflict
	case apperrors.KindUpstream:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if appErr, ok := apperrors.As(err); ok {
		if appErr.Err != nil {
			log.Printf("%s %s: %v", c.Method(), c.Path(), appErr.Err)
		}
		return fail(c, statusFor(appErr.Kind), appErr.Message, appErr.Details)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fail(c, fiberErr.Code, fiberErr.Message, nil)
	}

	log.Printf("%s %s: unhandled error: %v", c.Method(), c.Path(), err)
	return fail(c, http.StatusInternalServerError, "internal server error", nil)
}
