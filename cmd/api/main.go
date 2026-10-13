package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/7bryan/evermos-mini-api/internal/application/auth"
	"github.com/7bryan/evermos-mini-api/internal/application/region"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/handler"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/response"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/routes"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/validator"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/config"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/database"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/emsifa"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/security"
	mysqlrepo "github.com/7bryan/evermos-mini-api/internal/repository/mysql"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.NewMySQL(cfg.DB)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	if err := database.Migrate(db); err != nil {
		return err
	}

	// infrastructure
	hasher := security.NewPasswordHasher(security.DefaultBcryptCost)
	tokens := security.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiry)
	regionClient := emsifa.NewClient(cfg.Region.BaseURL, 5*time.Second, 24*time.Hour)

	// repositories
	userRepo := mysqlrepo.NewUserRepository(db)

	// services
	regionSvc := region.NewService(regionClient)
	authSvc, err := auth.NewService(userRepo, hasher, tokens, regionSvc)
	if err != nil {
		return err
	}

	// delivery
	val := validator.New()
	app := fiber.New(fiber.Config{
		AppName:      "evermos-mini-api",
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(recover.New())

	routes.Register(app, routes.Deps{
		Health: handler.Health(sqlDB),
		Auth:   handler.NewAuthHandler(authSvc, val),
		Region: handler.NewRegionHandler(regionSvc),
	})

	listenErr := make(chan error, 1)
	go func() { listenErr <- app.Listen(":" + cfg.App.Port) }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		return err
	case <-stop:
		log.Println("shutting down")
	}
	return app.ShutdownWithTimeout(10 * time.Second)
}
