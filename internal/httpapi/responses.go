package httpapi

type ExecuteScrapeResponse struct {
	TotalCount  int    `json:"totalCount"`
	NewCount    int    `json:"newCount"`
	StartedAt   int64  `json:"startedAt"`
	CompletedAt int64  `json:"completedAt"`
	ErrorKind   string `json:"errorKind,omitempty"`
	Error       string `json:"error,omitempty"`
}
