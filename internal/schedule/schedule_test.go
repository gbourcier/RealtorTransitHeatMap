package schedule

import (
	"errors"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestValidateScrapeParams(t *testing.T) {
	poly := ptr("MULTIPOLYGON (((0 0, 1 0, 1 1, 0 0)))")
	tests := []struct {
		name    string
		params  ScrapeParams
		wantErr error
	}{
		{
			name:    "polygon required",
			params:  ScrapeParams{},
			wantErr: ErrPolygonRequired,
		},
		{
			name:    "empty polygon rejected",
			params:  ScrapeParams{PolygonWKT: ptr("")},
			wantErr: ErrPolygonRequired,
		},
		{
			name:    "invalid building type",
			params:  ScrapeParams{PolygonWKT: poly, BuildingTypeID: ptr(99)},
			wantErr: ErrInvalidBuildingType,
		},
		{
			name:    "malformed bed range",
			params:  ScrapeParams{PolygonWKT: poly, BedRange: ptr("1+")},
			wantErr: ErrInvalidRange,
		},
		{
			name:    "price min greater than max",
			params:  ScrapeParams{PolygonWKT: poly, PriceMin: ptr(500), PriceMax: ptr(100)},
			wantErr: ErrInvalidPrice,
		},
		{
			name: "happy path",
			params: ScrapeParams{
				PolygonWKT:     poly,
				BuildingTypeID: ptr(17),
				BedRange:       ptr("2-0"),
				BathRange:      ptr(""),
				PriceMin:       ptr(100),
				PriceMax:       ptr(500),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScrapeParams(tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateScrapeParams = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
