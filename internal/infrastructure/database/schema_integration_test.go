//go:build integration

package database_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/7bryan/evermos-mini-api/internal/domain"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/config"
	"github.com/7bryan/evermos-mini-api/internal/infrastructure/database"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setup(t *testing.T) *gorm.DB {
	t.Helper()
	name := os.Getenv("TEST_DB_NAME")
	if name == "" {
		t.Skip("TEST_DB_NAME not set")
	}
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("refusing to run: TEST_DB_NAME %q must end with _test (tables get dropped)", name)
	}
	db, err := database.NewMySQL(config.DatabaseConfig{
		Host:     env("TEST_DB_HOST", "127.0.0.1"),
		Port:     env("TEST_DB_PORT", "3306"),
		User:     env("TEST_DB_USER", "evermos"),
		Password: os.Getenv("TEST_DB_PASSWORD"),
		Name:     name,
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
	return db
}

func mustCreate(t *testing.T, db *gorm.DB, v any) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create %T: %v", v, err)
	}
}

func userWith(email, phone string) *domain.User {
	return &domain.User{
		Name: "Test", PasswordHash: "x", Phone: phone, Email: email,
		BirthDate:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Occupation: "dev", ProvinceID: "11", CityID: "1101",
	}
}

type fixture struct {
	user     *domain.User
	store    *domain.Store
	category *domain.Category
	product  *domain.Product
}

func newProduct(t *testing.T, db *gorm.DB) fixture {
	t.Helper()
	u := userWith("owner@example.com", "0800000001")
	mustCreate(t, db, u)
	s := &domain.Store{UserID: u.ID, Name: "Store"}
	mustCreate(t, db, s)
	c := &domain.Category{Name: "Shoes"}
	mustCreate(t, db, c)
	p := &domain.Product{
		StoreID: s.ID, CategoryID: c.ID, Name: "Sneaker", Slug: "sneaker",
		ResellerPrice: 100, ConsumerPrice: 150, Stock: 3, Description: "d",
	}
	mustCreate(t, db, p)
	return fixture{u, s, c, p}
}

func newTransaction(t *testing.T, db *gorm.DB, f fixture) (*domain.Address, *domain.Transaction) {
	t.Helper()
	addr := &domain.Address{UserID: f.user.ID, Title: "Home", ReceiverName: "A", Phone: "0811", Detail: "Street 1"}
	mustCreate(t, db, addr)
	log := &domain.ProductLog{
		ProductID: f.product.ID, StoreID: f.store.ID, CategoryID: f.category.ID,
		Name: f.product.Name, Slug: f.product.Slug, ResellerPrice: 100, ConsumerPrice: 150,
		Description: "d", Photos: []domain.ProductLogPhoto{{URL: "a.png"}},
	}
	mustCreate(t, db, log)
	trx := &domain.Transaction{
		UserID: f.user.ID, AddressID: addr.ID, InvoiceCode: "INV-1", TotalPrice: 150, PaymentMethod: "bank",
		Items: []domain.TransactionItem{{ProductLogID: log.ID, Quantity: 1, TotalPrice: 150}},
	}
	mustCreate(t, db, trx)
	return addr, trx
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := setup(t)
	if err := database.Migrate(db); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}
}

func TestEmailAndPhoneMustBeUnique(t *testing.T) {
	db := setup(t)
	mustCreate(t, db, userWith("a@example.com", "0801"))
	if err := db.Create(userWith("a@example.com", "0802")).Error; err == nil {
		t.Fatal("duplicate email accepted")
	}
	if err := db.Create(userWith("b@example.com", "0801")).Error; err == nil {
		t.Fatal("duplicate phone accepted")
	}
}

func TestOneStorePerUser(t *testing.T) {
	db := setup(t)
	u := userWith("a@example.com", "0801")
	mustCreate(t, db, u)
	mustCreate(t, db, &domain.Store{UserID: u.ID, Name: "one"})
	if err := db.Create(&domain.Store{UserID: u.ID, Name: "two"}).Error; err == nil {
		t.Fatal("second store for same user accepted")
	}
}

func TestIsAdminDefaultsToFalse(t *testing.T) {
	db := setup(t)
	u := userWith("a@example.com", "0801")
	mustCreate(t, db, u)
	var got domain.User
	if err := db.First(&got, u.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.IsAdmin {
		t.Fatal("new user must not be admin")
	}
}

func TestStockCannotGoNegative(t *testing.T) {
	db := setup(t)
	f := newProduct(t, db) // stock = 3
	res := db.Model(&domain.Product{}).Where("id = ?", f.product.ID).
		Update("stock", gorm.Expr("stock - ?", 5))
	if res.Error == nil {
		t.Fatal("stock went below zero")
	}
}

func TestCategoryInUseCannotBeDeleted(t *testing.T) {
	db := setup(t)
	f := newProduct(t, db)
	if err := db.Delete(&domain.Category{}, f.category.ID).Error; err == nil {
		t.Fatal("category used by a product was deleted")
	}
}

func TestAddressUsedByTransactionCannotBeDeleted(t *testing.T) {
	db := setup(t)
	f := newProduct(t, db)
	addr, _ := newTransaction(t, db, f)
	if err := db.Delete(&domain.Address{}, addr.ID).Error; err == nil {
		t.Fatal("address used by a transaction was deleted")
	}
}

func TestDeletingProductKeepsTransactionHistory(t *testing.T) {
	db := setup(t)
	f := newProduct(t, db)
	_, trx := newTransaction(t, db, f)

	if err := db.Delete(&domain.Product{}, f.product.ID).Error; err != nil {
		t.Fatalf("product deletion failed: %v", err)
	}
	var got domain.Transaction
	if err := db.Preload("Items.ProductLog.Photos").First(&got, trx.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].ProductLog.Name != "Sneaker" || len(got.Items[0].ProductLog.Photos) != 1 {
		t.Fatalf("history lost: %+v", got.Items)
	}
}

func TestQuantityMustBePositive(t *testing.T) { // needs MySQL >= 8.0.16
	db := setup(t)
	f := newProduct(t, db)
	_, trx := newTransaction(t, db, f)
	err := db.Create(&domain.TransactionItem{
		TransactionID: trx.ID, ProductLogID: trx.Items[0].ProductLogID, Quantity: 0, TotalPrice: 0,
	}).Error
	if err == nil {
		t.Fatal("quantity 0 accepted")
	}
}
