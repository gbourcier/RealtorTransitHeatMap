package worker

import "time"

type job struct {
	reply chan ScrapeResult
}

// ScrapeResult is the outcome of a single scrape execution. ErrorKind/Error
// are populated only when Status indicates failure.
type ScrapeResult struct {
	Status      string
	StartedAt   time.Time
	CompletedAt time.Time
	TotalCount  int
	NewCount    int
	ErrorKind   string
	Error       string
}
