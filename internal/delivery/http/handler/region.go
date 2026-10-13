package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/application/region"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/response"
)

type RegionService interface {
	Provinces(ctx context.Context) ([]region.Province, error)
	Province(ctx context.Context, id string) (*region.Province, error)
	Cities(ctx context.Context, provinceID string) ([]region.City, error)
	City(ctx context.Context, id string) (*region.City, error)
}

type provinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cityResponse struct {
	ID         string `json:"id"`
	ProvinceID string `json:"province_id"`
	Name       string `json:"name"`
}

type RegionHandler struct {
	svc RegionService
}

func NewRegionHandler(svc RegionService) *RegionHandler {
	return &RegionHandler{svc: svc}
}

func (h *RegionHandler) ListProvinces(c *fiber.Ctx) error {
	list, err := h.svc.Provinces(c.UserContext())
	if err != nil {
		return err
	}
	out := make([]provinceResponse, 0, len(list))
	for _, p := range list {
		out = append(out, provinceResponse{ID: p.ID, Name: p.Name})
	}
	return response.OK(c, "provinces retrieved", out)
}

func (h *RegionHandler) DetailProvince(c *fiber.Ctx) error {
	p, err := h.svc.Province(c.UserContext(), c.Params("prov_id"))
	if err != nil {
		return err
	}
	return response.OK(c, "province retrieved", provinceResponse{ID: p.ID, Name: p.Name})
}

func (h *RegionHandler) ListCities(c *fiber.Ctx) error {
	list, err := h.svc.Cities(c.UserContext(), c.Params("prov_id"))
	if err != nil {
		return err
	}
	out := make([]cityResponse, 0, len(list))
	for _, v := range list {
		out = append(out, cityResponse{ID: v.ID, ProvinceID: v.ProvinceID, Name: v.Name})
	}
	return response.OK(c, "cities retrieved", out)
}

func (h *RegionHandler) DetailCity(c *fiber.Ctx) error {
	v, err := h.svc.City(c.UserContext(), c.Params("city_id"))
	if err != nil {
		return err
	}
	return response.OK(c, "city retrieved", cityResponse{ID: v.ID, ProvinceID: v.ProvinceID, Name: v.Name})
}
