package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/7bryan/evermos-mini-api/internal/application/auth"
	"github.com/7bryan/evermos-mini-api/internal/application/region"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/response"
	"github.com/7bryan/evermos-mini-api/internal/delivery/http/validator"
	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type fakeAuth struct {
	registered []auth.RegisterRequest
	loginErr   error
}

func (f *fakeAuth) Register(_ context.Context, r auth.RegisterRequest) error {
	f.registered = append(f.registered, r)
	return nil
}

func (f *fakeAuth) Login(context.Context, auth.LoginRequest) (*auth.LoginResponse, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return &auth.LoginResponse{Name: "Budi", Token: "tok"}, nil
}

type fakeRegion struct{}

func (fakeRegion) Provinces(context.Context) ([]region.Province, error) {
	return []region.Province{{ID: "11", Name: "ACEH"}}, nil
}
func (fakeRegion) Province(_ context.Context, id string) (*region.Province, error) {
	if id != "11" {
		return nil, apperrors.NotFound("province not found")
	}
	return &region.Province{ID: "11", Name: "ACEH"}, nil
}
func (fakeRegion) Cities(context.Context, string) ([]region.City, error) {
	return []region.City{{ID: "1101", ProvinceID: "11", Name: "SIMEULUE"}}, nil
}
func (fakeRegion) City(context.Context, string) (*region.City, error) {
	return &region.City{ID: "1101", ProvinceID: "11", Name: "SIMEULUE"}, nil
}

func newApp(a *fakeAuth) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	ah := NewAuthHandler(a, validator.New())
	rh := NewRegionHandler(fakeRegion{})
	app.Post("/register", ah.Register)
	app.Post("/login", ah.Login)
	app.Get("/provinces", rh.ListProvinces)
	app.Get("/province/:prov_id", rh.DetailProvince)
	app.Get("/cities/:prov_id", rh.ListCities)
	return app
}

func do(t *testing.T, app *fiber.App, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	return resp.StatusCode, out
}

func TestRegisterEmptyBodyListsMissingFields(t *testing.T) {
	a := &fakeAuth{}
	code, body := do(t, newApp(a), "POST", "/register", `{}`)
	if code != 400 || body["status"] != false {
		t.Fatalf("code=%d body=%v", code, body)
	}
	details, _ := json.Marshal(body["errors"])
	for _, field := range []string{"nama", "kata_sandi", "no_telp", "tanggal_Lahir", "pekerjaan", "email", "id_provinsi", "id_kota"} {
		if !strings.Contains(string(details), field+" is required") {
			t.Errorf("missing detail for %s: %s", field, details)
		}
	}
	if len(a.registered) != 0 {
		t.Fatal("service called with invalid input")
	}
}

func TestRegisterMalformedJSON(t *testing.T) {
	code, _ := do(t, newApp(&fakeAuth{}), "POST", "/register", `{"nama":`)
	if code != 400 {
		t.Fatalf("code = %d, want 400", code)
	}
}

func TestRegisterSuccess(t *testing.T) {
	a := &fakeAuth{}
	body := `{"nama":"Budi","kata_sandi":"secret123","no_telp":"081234567890","tanggal_Lahir":"17/08/1990",
		"pekerjaan":"Engineer","email":"budi@example.com","id_provinsi":"11","id_kota":"1101"}`
	code, out := do(t, newApp(a), "POST", "/register", body)
	if code != 201 || out["status"] != true {
		t.Fatalf("code=%d out=%v", code, out)
	}
	if len(a.registered) != 1 || a.registered[0].Email != "budi@example.com" {
		t.Fatalf("service input = %+v", a.registered)
	}
}

func TestRegisterCannotSetAdmin(t *testing.T) {
	a := &fakeAuth{}
	body := `{"nama":"Budi","kata_sandi":"secret123","no_telp":"081234567890","tanggal_Lahir":"17/08/1990",
		"pekerjaan":"Engineer","email":"budi@example.com","id_provinsi":"11","id_kota":"1101","is_admin":true}`
	if code, _ := do(t, newApp(a), "POST", "/register", body); code != 201 {
		t.Fatalf("code = %d", code)
	}
	// RegisterRequest has no IsAdmin field, so the extra key is simply ignored.
	if _, has := interface{}(a.registered[0]).(interface{ IsAdmin() bool }); has {
		t.Fatal("request type exposes an admin flag")
	}
}

func TestLoginSuccessAndFailure(t *testing.T) {
	code, out := do(t, newApp(&fakeAuth{}), "POST", "/login", `{"no_telp":"0812","kata_sandi":"x"}`)
	data, _ := out["data"].(map[string]any)
	if code != 200 || data["token"] != "tok" {
		t.Fatalf("code=%d out=%v", code, out)
	}

	failing := &fakeAuth{loginErr: apperrors.Unauthorized("invalid phone number or password")}
	if code, _ := do(t, newApp(failing), "POST", "/login", `{"no_telp":"0812","kata_sandi":"x"}`); code != 401 {
		t.Fatalf("code = %d, want 401", code)
	}
}

func TestLoginValidation(t *testing.T) {
	if code, _ := do(t, newApp(&fakeAuth{}), "POST", "/login", `{"no_telp":"abc","kata_sandi":""}`); code != 400 {
		t.Fatalf("code = %d, want 400", code)
	}
}

func TestRegionHandlers(t *testing.T) {
	app := newApp(&fakeAuth{})

	code, out := do(t, app, "GET", "/provinces", "")
	list, _ := out["data"].([]any)
	first, _ := list[0].(map[string]any)
	if code != 200 || first["id"] != "11" || first["name"] != "ACEH" {
		t.Fatalf("provinces: code=%d out=%v", code, out)
	}

	code, out = do(t, app, "GET", "/cities/11", "")
	list, _ = out["data"].([]any)
	first, _ = list[0].(map[string]any)
	if code != 200 || first["province_id"] != "11" {
		t.Fatalf("cities: code=%d out=%v", code, out)
	}

	if code, _ := do(t, app, "GET", "/province/99", ""); code != 404 {
		t.Fatalf("unknown province code = %d, want 404", code)
	}
}
