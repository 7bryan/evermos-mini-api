package emsifa

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/7bryan/evermos-mini-api/internal/application/region"
)

const (
	provincesJSON = `[{"id":"11","name":"ACEH"},{"id":"12","name":"SUMATERA UTARA"}]`
	citiesJSON    = `[{"id":"1101","province_id":"11","name":"KABUPATEN SIMEULUE"},{"id":"1102","province_id":"11","name":"KABUPATEN ACEH SINGKIL"}]`
	cityJSON      = `{"id":"1101","province_id":"11","name":"KABUPATEN SIMEULUE"}`
)

type fakeAPI struct {
	*httptest.Server
	mu    sync.Mutex
	paths []string
}

func (f *fakeAPI) calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.paths...)
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()
	f := &fakeAPI{}
	serve := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f.mu.Lock()
			f.paths = append(f.paths, r.URL.Path)
			f.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/provinces.json", serve(provincesJSON))
	mux.HandleFunc("/regencies/11.json", serve(citiesJSON))
	mux.HandleFunc("/regency/1101.json", serve(cityJSON))
	// every other path gets the mux's own 404
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func newClient(f *fakeAPI) *Client {
	return NewClient(f.URL, time.Second, time.Hour)
}

func TestProvincesAreCached(t *testing.T) {
	api := newFakeAPI(t)
	c := newClient(api)
	for i := 0; i < 3; i++ {
		got, err := c.Provinces(context.Background())
		if err != nil || len(got) != 2 || got[0].Name != "ACEH" {
			t.Fatalf("got=%v err=%v", got, err)
		}
	}
	if n := len(api.calls()); n != 1 {
		t.Fatalf("upstream calls = %d, want 1", n)
	}
}

func TestCacheExpires(t *testing.T) {
	api := newFakeAPI(t)
	c := newClient(api)
	now := time.Now()
	c.now = func() time.Time { return now }

	_, _ = c.Provinces(context.Background())
	now = now.Add(2 * time.Hour)
	_, _ = c.Provinces(context.Background())

	if n := len(api.calls()); n != 2 {
		t.Fatalf("upstream calls = %d, want 2", n)
	}
}

func TestServesStaleDataWhenUpstreamFails(t *testing.T) {
	var mu sync.Mutex
	down := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		isDown := down
		mu.Unlock()
		if isDown {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(provincesJSON))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, time.Second, time.Hour)
	now := time.Now()
	c.now = func() time.Time { return now }

	if _, err := c.Provinces(context.Background()); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	down = true
	mu.Unlock()
	now = now.Add(2 * time.Hour)

	got, err := c.Provinces(context.Background())
	if err != nil || len(got) != 2 {
		t.Fatalf("stale data not served: got=%v err=%v", got, err)
	}
}

func TestCachedSliceCannotBeMutatedByCaller(t *testing.T) {
	c := newClient(newFakeAPI(t))
	first, _ := c.Provinces(context.Background())
	first[0].Name = "HACKED"
	second, _ := c.Provinces(context.Background())
	if second[0].Name != "ACEH" {
		t.Fatal("cache was mutated through a returned slice")
	}
}

func TestProvinceLookup(t *testing.T) {
	c := newClient(newFakeAPI(t))
	p, err := c.Province(context.Background(), "11")
	if err != nil || p.Name != "ACEH" {
		t.Fatalf("p=%v err=%v", p, err)
	}
	if _, err := c.Province(context.Background(), "99"); !errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCities(t *testing.T) {
	c := newClient(newFakeAPI(t))
	got, err := c.Cities(context.Background(), "11")
	if err != nil || len(got) != 2 || got[0].ProvinceID != "11" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestCitiesOfUnknownProvinceNeverCallsRegencies(t *testing.T) {
	api := newFakeAPI(t)
	c := newClient(api)
	if _, err := c.Cities(context.Background(), "99"); !errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	for _, p := range api.calls() {
		if p != "/provinces.json" {
			t.Fatalf("unexpected upstream call: %s", p)
		}
	}
}

func TestCityLookup(t *testing.T) {
	c := newClient(newFakeAPI(t))
	city, err := c.City(context.Background(), "1101")
	if err != nil || city.Name != "KABUPATEN SIMEULUE" || city.ProvinceID != "11" {
		t.Fatalf("city=%v err=%v", city, err)
	}
	if _, err := c.City(context.Background(), "9999"); !errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMalformedIDsNeverReachUpstream(t *testing.T) {
	api := newFakeAPI(t)
	c := newClient(api)
	for _, id := range []string{"", "abc", "../etc/passwd", "11/../12", "1 1", "11?x=1", "12345678901"} {
		if _, err := c.Province(context.Background(), id); !errors.Is(err, region.ErrNotFound) {
			t.Fatalf("Province(%q) err = %v", id, err)
		}
		if _, err := c.Cities(context.Background(), id); !errors.Is(err, region.ErrNotFound) {
			t.Fatalf("Cities(%q) err = %v", id, err)
		}
		if _, err := c.City(context.Background(), id); !errors.Is(err, region.ErrNotFound) {
			t.Fatalf("City(%q) err = %v", id, err)
		}
	}
	if n := len(api.calls()); n != 0 {
		t.Fatalf("malformed IDs caused %d upstream calls", n)
	}
}

func TestUpstreamErrorIsNotReportedAsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, time.Second, time.Hour).Provinces(context.Background())
	if err == nil || errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want a non-NotFound error", err)
	}
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write([]byte(provincesJSON))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, 50*time.Millisecond, time.Hour).Provinces(context.Background())
	if err == nil || errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want a timeout error", err)
	}
}

func TestCanceledContext(t *testing.T) {
	c := newClient(newFakeAPI(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Provinces(ctx); err == nil {
		t.Fatal("expected an error for a canceled context")
	}
}

func TestInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, time.Second, time.Hour).Provinces(context.Background())
	if err == nil || errors.Is(err, region.ErrNotFound) {
		t.Fatalf("err = %v, want a decode error", err)
	}
}
