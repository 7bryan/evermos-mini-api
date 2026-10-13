//go:build integration

package mysql_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/7bryan/evermos-mini-api/internal/domain"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/config"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/database"
	mysqlrepo "github.com/7bryan/evermos-mini-api/internal/repository/mysql"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setup(t *testing.T) *mysqlrepo.UserRepository {
	t.Helper()
	name := os.Getenv("TEST_DB_NAME")
	if name == "" {
		t.Skip("TEST_DB_NAME not set")
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("refusing to run: TEST_DB_NAME %q must end with _test", name)
	}
	db, err := database.NewMySQL(config.DatabaseConfig{
		Host: env("TEST_DB_HOST", "127.0.0.1"), Port: env("TEST_DB_PORT", "3306"),
		User: env("TEST_DB_USER", "evermos"), Password: os.Getenv("TEST_DB_PASSWORD"), Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })

	models := database.Models()
	for i := len(models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(models[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return mysqlrepo.NewUserRepository(db)
}

func user(email, phone string) *domain.User {
	return &domain.User{
		Name: "Test", PasswordHash: "x", Phone: phone, Email: email,
		BirthDate:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Occupation: "dev", ProvinceID: "11", CityID: "1101",
	}
}

var _ = gorm.ErrRecordNotFound

func TestCreateWithStoreAndFind(t *testing.T) {
	repo := setup(t)
	u := user("a@example.com", "0801")
	if err := repo.CreateWithStore(context.Background(), u, &domain.Store{Name: "Toko A"}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByPhone(context.Background(), "0801")
	if err != nil || got.ID != u.ID || got.IsAdmin {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	isAdmin, err := repo.IsAdmin(context.Background(), u.ID)
	if err != nil || isAdmin {
		t.Fatalf("isAdmin=%v err=%v", isAdmin, err)
	}
	if _, err := repo.IsAdmin(context.Background(), 9999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if _, err := repo.FindByPhone(context.Background(), "0000"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDuplicatePhoneAndEmailAreTranslated(t *testing.T) {
	repo := setup(t)
	ctx := context.Background()
	if err := repo.CreateWithStore(ctx, user("a@example.com", "0801"), &domain.Store{Name: "A"}); err != nil {
		t.Fatal(err)
	}
	err := repo.CreateWithStore(ctx, user("other@example.com", "0801"), &domain.Store{Name: "B"})
	if !errors.Is(err, domain.ErrDuplicatePhone) {
		t.Fatalf("err = %v, want ErrDuplicatePhone", err)
	}
	err = repo.CreateWithStore(ctx, user("a@example.com", "0802"), &domain.Store{Name: "C"})
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Fatalf("err = %v, want ErrDuplicateEmail", err)
	}
}

func TestDuplicateKeyIsMatchedByIndexNotValue(t *testing.T) {
	repo := setup(t)
	ctx := context.Background()
	if err := repo.CreateWithStore(ctx, user("phone@example.com", "0801"), &domain.Store{Name: "A"}); err != nil {
		t.Fatal(err)
	}
	err := repo.CreateWithStore(ctx, user("phone@example.com", "0802"), &domain.Store{Name: "B"})
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Fatalf("err = %v, want ErrDuplicateEmail", err)
	}
}

func TestFailedStoreInsertRollsBackUser(t *testing.T) {
	repo := setup(t)
	ctx := context.Background()
	err := repo.CreateWithStore(ctx, user("a@example.com", "0801"), &domain.Store{Name: strings.Repeat("x", 300)})
	if err == nil {
		t.Fatal("oversized store name accepted")
	}
	if _, err := repo.FindByPhone(ctx, "0801"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("user survived a failed store insert: %v", err)
	}
}
