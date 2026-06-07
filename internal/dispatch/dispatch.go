package dispatch

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs/refresh"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/realtor"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scrape"
	"github.com/google/uuid"
)

var (
	ErrUnknownJobType = errors.New("dispatch: unknown job type")
	ErrBusy           = errors.New("dispatch: worker busy")
)

// IsBusy reports whether err indicates a worker was already in progress.
// Recognises both dispatch.ErrBusy and the underlying worker sentinels.
func (d *Dispatcher) IsBusy(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrBusy) ||
		errors.Is(err, scrape.ErrBusy) ||
		errors.Is(err, refresh.ErrBusy)
}

type ScrapeTrigger interface {
	StartScrapeForSchedule(scheduleID uuid.UUID, scheduleName string, params realtor.SearchParams) (uuid.UUID, error)
}

type GtfsRefreshTrigger interface {
	StartForSchedule(scheduleID uuid.UUID) (uuid.UUID, error)
}

type Dispatcher struct {
	scrape  ScrapeTrigger
	refresh GtfsRefreshTrigger
}

func New(scrape ScrapeTrigger, refresh GtfsRefreshTrigger) *Dispatcher {
	return &Dispatcher{scrape: scrape, refresh: refresh}
}

func (d *Dispatcher) Dispatch(s schedule.Schedule) (uuid.UUID, error) {
	switch s.JobType {
	case schedule.JobTypeScrapeRealtor:
		return d.scrape.StartScrapeForSchedule(s.ID, s.Name, searchParamsFromSchedule(s))
	case schedule.JobTypeRefreshGtfs:
		return d.refresh.StartForSchedule(s.ID)
	default:
		return uuid.Nil, fmt.Errorf("%w: %q", ErrUnknownJobType, s.JobType)
	}
}

func searchParamsFromSchedule(s schedule.Schedule) realtor.SearchParams {
	return realtor.SearchParams{
		BuildingTypeID: intPtrToStr(s.BuildingTypeID),
		BedRange:       strPtr(s.BedRange),
		BathRange:      strPtr(s.BathRange),
		PriceMin:       intPtrToStr(s.PriceMin),
		PriceMax:       intPtrToStr(s.PriceMax),
		PolygonWKT:     strPtr(s.PolygonWKT),
	}
}

func intPtrToStr(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
