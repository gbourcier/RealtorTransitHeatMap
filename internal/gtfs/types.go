package gtfs

import "time"

type FeedSource struct {
	Agency  string
	ZipPath string
}

type Stop struct {
	Agency    string
	StopID    string
	Name      string
	Latitude  float64
	Longitude float64
}

type Result struct {
	Stop
	CommuteSecondsToTarget *int
}

type stopTime struct {
	stopIdx int
	arrSec  int
	depSec  int
	seq     int
}

type trip struct {
	serviceID string
	stopTimes []stopTime
}

type connection struct {
	depStop int
	arrStop int
	depSec  int
	arrSec  int
}

type transfer struct {
	fromStop int
	toStop   int
	seconds  int
}

type service struct {
	weekdays    [7]bool
	startDate   time.Time
	endDate     time.Time
	addedDates  map[string]bool
	removedDays map[string]bool
}

type feed struct {
	agency    string
	stops     []Stop
	trips     map[string]*trip
	services  map[string]*service
	transfers []rawTransfer
}

type rawTransfer struct {
	fromStopID string
	toStopID   string
	minSec     int
}
