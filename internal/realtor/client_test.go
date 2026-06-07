package realtor

import (
	"context"
	"testing"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

func TestFetchPricesMock(t *testing.T) {
	c := NewClient(config.RealtorConfig{Mock: true})
	obs, err := c.FetchPrices(context.Background(), SearchParams{})
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
				Building: building{
					BathroomTotal: "2",
					Bedrooms:      "3",
					SizeInterior:  "1115.14 sqft",
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

func TestParseSqft(t *testing.T) {
	cases := map[string]float64{
		"":             0,
		"1115.14 sqft": 1115.14,
		"827 sqft":     827,
		"  500 sqft  ": 500,
		"not-a-number": 0,
	}
	for in, want := range cases {
		got := parseSqft(in)
		if got != want {
			t.Errorf("parseSqft(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDecodeObservationsBuildingFields(t *testing.T) {
	results := []listingResult{{
		MlsNumber:  1,
		Individual: []individual{{Organization: organization{OrganizationId: 1}}},
		Building: building{
			BathroomTotal: "2",
			Bedrooms:      "3",
			SizeInterior:  "1115.14 sqft",
			Type:          "Triplex",
		},
	}}
	got := decodeObservations(results)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	l := got[0].Listing
	if l.BedroomCount != 3 {
		t.Errorf("BedroomCount = %d, want 3", l.BedroomCount)
	}
	if l.BathroomCount != 2 {
		t.Errorf("BathroomCount = %d, want 2", l.BathroomCount)
	}
	if l.InteriorAreaSqft != 1115.14 {
		t.Errorf("InteriorAreaSqft = %v, want 1115.14", l.InteriorAreaSqft)
	}
	if l.BuildingType != listing.BuildingTypeTriplex {
		t.Errorf("BuildingType = %d, want %d", l.BuildingType, listing.BuildingTypeTriplex)
	}
}

func TestParseBuildingType(t *testing.T) {
	cases := map[string]listing.BuildingType{
		"":          listing.BuildingTypeHouse,
		"House":     listing.BuildingTypeHouse,
		"Detached":  listing.BuildingTypeHouse,
		"Duplex":    listing.BuildingTypeDuplex,
		"duplex":    listing.BuildingTypeDuplex,
		" Triplex ": listing.BuildingTypeTriplex,
		"Fourplex":  listing.BuildingTypeFourplex,
	}
	for in, want := range cases {
		got := parseBuildingType(in)
		if got != want {
			t.Errorf("parseBuildingType(%q) = %d, want %d", in, got, want)
		}
	}
}
