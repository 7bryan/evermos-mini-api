package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/delivery/http/handler"
)

type Deps struct {
	Health fiber.Handler
	Auth   *handler.AuthHandler
	Region *handler.RegionHandler
}

func Register(app *fiber.App, d Deps) {
	app.Get("/health", d.Health)

	api := app.Group("/api/v1")

	authGroup := api.Group("/auth")
	authGroup.Post("/register", d.Auth.Register)
	authGroup.Post("/login", d.Auth.Login)

	prov := api.Group("/provcity")
	prov.Get("/listprovincies", d.Region.ListProvinces)
	prov.Get("/listcities/:prov_id", d.Region.ListCities)
	prov.Get("/detailprovince/:prov_id", d.Region.DetailProvince)
	prov.Get("/detailcity/:city_id", d.Region.DetailCity)
}
