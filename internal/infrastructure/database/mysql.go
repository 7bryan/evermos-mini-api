package database

import (
	"context"
	"fmt"
	"net"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/7bryan/evermos-mini-api/internal/infrastructure/config"
)

func NewMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsnConfig := driver.NewConfig()
	dsnConfig.User = cfg.User
	dsnConfig.Passwd = cfg.Password
	dsnConfig.Net = "tcp"
	dsnConfig.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	dsnConfig.DBName = cfg.Name
	dsnConfig.ParseTime = true
	dsnConfig.Collation = "utf8mb4_unicode_ci"

	db, err := gorm.Open(gormmysql.Open(dsnConfig.FormatDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
