package api

import (
	"errors"
	"net/http"
	"strconv"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

// parsePage reads ?limit= and ?offset= and applies defaults / bounds. Returns
// a 400-ready error message if a value is invalid.
func parsePage(r *http.Request) (limit, offset int, err error) {
	limit = defaultPageLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 1 {
			return 0, 0, errors.New("invalid 'limit' query param")
		}
		if n > maxPageLimit {
			n = maxPageLimit
		}
		limit = n
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 0 {
			return 0, 0, errors.New("invalid 'offset' query param")
		}
		offset = n
	}
	return limit, offset, nil
}
