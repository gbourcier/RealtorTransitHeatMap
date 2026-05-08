package realtor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

const (
	defaultBaseURL = "https://api2.realtor.ca"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	http      *http.Client
	baseURL   string
	userAgent string
}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: defaultTimeout,
		},
		baseURL:   defaultBaseURL,
		userAgent: defaultUserAgent,
	}
}

func (c *Client) FetchPrices(ctx context.Context, criteria listing.FetchCriteria) ([]listing.Listing, error) {
	slog.Info("fetching prices", "city", criteria.City)
	// TODO: replace with a real call to realtor.ca and parse the response.
	return []listing.Listing{
		{
			ID:      "stub-1",
			Price:   750_000,
			Address: "123 Fake St, " + criteria.City,
		},
	}, nil
}
