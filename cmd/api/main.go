package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/infrastructure/config"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/database"
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

	app := fiber.New(fiber.Config{AppName: "evermos-mini-api"})

	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
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
