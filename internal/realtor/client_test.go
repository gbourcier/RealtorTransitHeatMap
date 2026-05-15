package realtor

import (
	"context"
	"testing"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
)

func TestFetchPricesMock(t *testing.T) {
	c := NewClient(config.RealtorConfig{Mock: true})
	obs, err := c.FetchPrices(context.Background())
	if err != nil {
		t.Fatalf("FetchPrices: %v", err)
	}
	if len(obs) == 0 {
		t.Fatal("expected at least one observation from mock response")
	}
	for i, o := range obs {
		if o.Listing.MLS == 0 {
			t.Errorf("obs[%d]: MLS is zero", i)
		}
		if o.Listing.Board == 0 {
			t.Errorf("obs[%d]: Board is zero", i)
		}
		if o.Price == 0 {
			t.Errorf("obs[%d]: Price is zero", i)
		}
		if o.Listing.Address == "" {
			t.Errorf("obs[%d]: Address is empty", i)
		}
	}
}

func TestDecodeObservations(t *testing.T) {
	tests := []struct {
		name    string
		results []listingResult
		want    int
	}{
		{
			name:    "empty input",
			results: nil,
			want:    0,
		},
		{
			name: "skips results without organization",
			results: []listingResult{
				{MlsNumber: 1, Individual: nil},
			},
			want: 0,
		},
		{
			name: "happy path",
			results: []listingResult{{
				MlsNumber: 42,
				StatusId:  "active",
				Property: property{
					PriceUnformattedValue: "500000",
					Address: address{
						AddressText: "123 Main St",
						Latitude:    "45.5",
						Longitude:   "-73.5",
					},
				},
				Individual: []individual{{Organization: organization{OrganizationId: 7}}},
			}},
			want: 1,
		},
		{
			name: "malformed numerics become zero, listing still emitted",
			results: []listingResult{{
				MlsNumber:  9,
				Individual: []individual{{Organization: organization{OrganizationId: 7}}},
				Property: property{
					PriceUnformattedValue: "not-a-number",
					Address:               address{Latitude: "x", Longitude: "y"},
				},
			}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeObservations(tt.results)
			if len(got) != tt.want {
				t.Fatalf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}
