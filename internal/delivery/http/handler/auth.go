package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/application/auth"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/response"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/validator"
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) error
	Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, error)
}

type AuthHandler struct {
	svc AuthService
	val *validator.Validator
}

func NewAuthHandler(svc AuthService, val *validator.Validator) *AuthHandler {
	return &AuthHandler{svc: svc, val: val}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req auth.RegisterRequest
	if err := bind(c, h.val, &req); err != nil {
		return err
	}
	if err := h.svc.Register(c.UserContext(), req); err != nil {
		return err
	}
	return response.Created(c, "registration succeeded", "Register Succeed")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := bind(c, h.val, &req); err != nil {
		return err
	}
	res, err := h.svc.Login(c.UserContext(), req)
	if err != nil {
		return err
	}
	return response.OK(c, "login succeeded", res)
}
