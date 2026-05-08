package realtor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

const (
	defaultTimeout = 30 * time.Second
	searchPath     = "/Listing.svc/AsyncPropertySearch_Post"
	pageDelay      = 2 * time.Second
	boardName      = "realtor.ca"
)

type Client struct {
	http    *http.Client
	baseURL string
	cfg     *config.Config
}

func NewClient(cfg *config.Config) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: defaultTimeout,
		},
		baseURL: cfg.RealtorBaseURL,
		cfg:     cfg,
	}
}

func (c *Client) FetchPrices(ctx context.Context, criteria listing.FetchCriteria) ([]listing.Listing, error) {
	slog.Info("fetching prices from realtor.ca")
	var all []listing.Listing
	page := 1
	for {
		batch, totalPages, err := c.fetchPage(ctx, criteria, page)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		all = append(all, batch...)
		slog.Info("fetched page", "page", page, "total_pages", totalPages, "count", len(batch))
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

func (c *Client) fetchPage(ctx context.Context, criteria listing.FetchCriteria, page int) ([]listing.Listing, int, error) {
	vals := searchValues(c.cfg, criteria, page)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+searchPath, strings.NewReader(vals.Encode()))
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
		return nil, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var parsed asyncPropertySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, 0, fmt.Errorf("decode response: %w", err)
	}

	listings := make([]listing.Listing, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		lat, _ := strconv.ParseFloat(r.Property.Address.Latitude, 64)
		lon, _ := strconv.ParseFloat(r.Property.Address.Longitude, 64)
		price, _ := strconv.ParseFloat(r.Property.PriceUnformattedValue, 64)
		listings = append(listings, listing.Listing{
			Board:     boardName,
			MLS:       r.MlsNumber,
			Latitude:  lat,
			Longitude: lon,
			Address:   r.Property.Address.AddressText,
			Status:    r.StatusId,
			Price:     price,
		})
	}

	return listings, parsed.Paging.TotalPages, nil
}
