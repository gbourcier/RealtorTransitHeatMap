package realtor

import "net/http"

const defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func (c *Client) defaultHeaders() http.Header {
	h := http.Header{}
	h.Set("User-Agent", c.userAgent)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Origin", "https://www.realtor.ca")
	h.Set("Referer", "https://www.realtor.ca/")
	return h
}
