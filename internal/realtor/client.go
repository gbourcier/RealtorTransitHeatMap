package realtor

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

var (
	ErrForbidden = errors.New("realtor: forbidden")
	ErrNullCount = errors.New("realtor: null count")
)

//go:embed mock/response.json
var mockResponseJSON []byte

const (
	defaultTimeout = 30 * time.Second
	searchPath     = "/Listing.svc/AsyncPropertySearch_Post"
	primeURL       = "https://www.realtor.ca/map"
	pageDelay      = 2 * time.Second
	maxErrBody     = 2048
	primeTTL       = 30 * time.Minute
)

type Client struct {
	http      *http.Client
	cfg       config.RealtorConfig
	lastPrime time.Time
}

func NewClient(cfg config.RealtorConfig) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{
			Jar:       jar,
			Transport: browserTransport(),
			Timeout:   defaultTimeout,
		},
		cfg: cfg,
	}
}

func (c *Client) fetchMock() ([]listing.Observation, error) {
	slog.Info("fetching prices from mock response")
	var parsed asyncPropertySearchResponse
	if err := json.Unmarshal(mockResponseJSON, &parsed); err != nil {
		return nil, fmt.Errorf("decode mock response: %w", err)
	}
	return decodeObservations(parsed.Results), nil
}

func (c *Client) FetchPrices(ctx context.Context) ([]listing.Observation, error) {
	if c.cfg.Mock {
		return c.fetchMock()
	}
	slog.Info("fetching prices from realtor.ca")
	if err := c.ensurePrimed(ctx); err != nil {
		return nil, fmt.Errorf("prime session: %w", err)
	}
	var all []listing.Observation
	page := 1
	for {
		batch, totalPages, err := c.fetchPage(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		all = append(all, batch...)
		slog.Info("fetched page", "page", page, "total_pages", totalPages, "count", len(batch))
		if page == 1 && totalPages == 0 && len(batch) == 0 {
			return nil, ErrNullCount
		}
		if page >= totalPages {
			break
		}
		page++
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pageDelay):
		}
	}
	return all, nil
}

func (c *Client) ensurePrimed(ctx context.Context) error {
	if time.Since(c.lastPrime) < primeTTL {
		return nil
	}
	if err := c.primeSession(ctx); err != nil {
		return err
	}
	c.lastPrime = time.Now()
	return nil
}

func (c *Client) primeSession(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, primeURL, nil)
	if err != nil {
		return err
	}
	for k, vv := range c.defaultHeaders() {
		for _, v := range vv {
			req.Header.Set(k, v)
		}
	}
	req.Header.Del("Origin")
	req.Header.Del("Content-Type")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("prime status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) fetchPage(ctx context.Context, page int) ([]listing.Observation, int, error) {
	vals := searchValues(c.cfg, page)
	encoded := vals.Encode()
	slog.Debug("realtor search request", "page", page, "body", encoded)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.BaseURL+searchPath, strings.NewReader(encoded))
	if err != nil {
		return nil, 0, err
	}
	for k, vv := range c.defaultHeaders() {
		for _, v := range vv {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBody))
		slog.Error("realtor search non-200",
			"status", resp.StatusCode,
			"page", page,
			"body", string(body))
		if resp.StatusCode == http.StatusForbidden {
			return nil, 0, ErrForbidden
		}
		return nil, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read body: %w", err)
	}
	var parsed asyncPropertySearchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		preview := body
		if len(preview) > maxErrBody {
			preview = preview[:maxErrBody]
		}
		slog.Error("realtor search decode failed",
			"page", page,
			"content_encoding", resp.Header.Get("Content-Encoding"),
			"content_type", resp.Header.Get("Content-Type"),
			"body_preview", string(preview))
		return nil, 0, fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Results) == 0 {
		preview := body
		if len(preview) > maxErrBody {
			preview = preview[:maxErrBody]
		}
		slog.Warn("realtor search returned no results",
			"page", page,
			"error_code", parsed.ErrorCode,
			"paging", parsed.Paging,
			"body_preview", string(preview))
	}

	return decodeObservations(parsed.Results), parsed.Paging.TotalPages, nil
}

func decodeObservations(results []listingResult) []listing.Observation {
	out := make([]listing.Observation, 0, len(results))
	for _, r := range results {
		if len(r.Individual) == 0 {
			continue
		}
		lat, _ := strconv.ParseFloat(r.Property.Address.Latitude, 64)
		lon, _ := strconv.ParseFloat(r.Property.Address.Longitude, 64)
		price, _ := strconv.ParseFloat(r.Property.PriceUnformattedValue, 64)
		out = append(out, listing.Observation{
			Listing: listing.Listing{
				Board:     r.Individual[0].Organization.OrganizationId,
				MLS:       r.MlsNumber,
				Latitude:  lat,
				Longitude: lon,
				Address:   r.Property.Address.AddressText,
				IsAvailable: r.StatusId == "1",
				Slug:      r.RelativeDetailsURL,
			},
			Price: price,
		})
	}
	return out
}
