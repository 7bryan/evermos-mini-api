package emsifa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/7bryan/evermos-mini-api/internal/application/region"
)

const (
	DefaultBaseURL = "https://emsifa.github.io/api-wilayah-indonesia/api"
	maxBodyBytes   = 2 << 20 // 2 MB
)

var idPattern = regexp.MustCompile(`^[0-9]{1,10}$`)

type provinceWire struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cityWire struct {
	ID         string `json:"id"`
	ProvinceID string `json:"province_id"`
	Name       string `json:"name"`
}

type item struct {
	value   any
	expires time.Time
}

type Client struct {
	baseURL string
	http    *http.Client
	ttl     time.Duration
	now     func() time.Time

	mu    sync.Mutex
	items map[string]item
}

var _ region.Provider = (*Client)(nil)

func NewClient(baseURL string, timeout, cacheTTL time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
		ttl:     cacheTTL,
		now:     time.Now,
		items:   make(map[string]item),
	}
}

func (c *Client) Provinces(ctx context.Context) ([]region.Province, error) {
	v, err := cached(c, "provinces", func() ([]region.Province, error) {
		var raw []provinceWire
		if err := c.get(ctx, "/provinces.json", &raw); err != nil {
			return nil, err
		}
		out := make([]region.Province, 0, len(raw))
		for _, p := range raw {
			out = append(out, region.Province{ID: p.ID, Name: p.Name})
		}
		return out, nil
	})
	return slices.Clone(v), err
}

func (c *Client) Province(ctx context.Context, id string) (*region.Province, error) {
	if !idPattern.MatchString(id) {
		return nil, region.ErrNotFound
	}
	list, err := c.Provinces(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, region.ErrNotFound
}

func (c *Client) Cities(ctx context.Context, provinceID string) ([]region.City, error) {
	if _, err := c.Province(ctx, provinceID); err != nil {
		return nil, err
	}
	v, err := cached(c, "cities:"+provinceID, func() ([]region.City, error) {
		var raw []cityWire
		if err := c.get(ctx, "/regencies/"+provinceID+".json", &raw); err != nil {
			return nil, err
		}
		out := make([]region.City, 0, len(raw))
		for _, r := range raw {
			out = append(out, region.City{ID: r.ID, ProvinceID: r.ProvinceID, Name: r.Name})
		}
		return out, nil
	})
	return slices.Clone(v), err
}

func (c *Client) City(ctx context.Context, id string) (*region.City, error) {
	if !idPattern.MatchString(id) {
		return nil, region.ErrNotFound
	}
	v, err := cached(c, "city:"+id, func() (region.City, error) {
		var raw cityWire
		if err := c.get(ctx, "/regency/"+id+".json", &raw); err != nil {
			return region.City{}, err
		}
		if raw.ID == "" {
			return region.City{}, errors.New("unexpected city payload: empty id")
		}
		return region.City{ID: raw.ID, ProvinceID: raw.ProvinceID, Name: raw.Name}, nil
	})
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build region request: %w", err)
	}
	req.Header.Set("User-Agent", "evermos-mini-api/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("region request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return region.ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("region upstream returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(out); err != nil {
		return fmt.Errorf("decode region response: %w", err)
	}
	return nil
}

func cached[T any](c *Client, key string, fetch func() (T, error)) (T, error) {
	c.mu.Lock()
	it, ok := c.items[key]
	c.mu.Unlock()
	if ok && c.now().Before(it.expires) {
		return it.value.(T), nil
	}

	v, err := fetch()
	if err != nil {
		if ok && !errors.Is(err, region.ErrNotFound) {
			return it.value.(T), nil
		}
		var zero T
		return zero, err
	}

	c.mu.Lock()
	c.items[key] = item{value: v, expires: c.now().Add(c.ttl)}
	c.mu.Unlock()
	return v, nil
}
