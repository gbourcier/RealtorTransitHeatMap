package realtor

import "net/http"

const defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0"

func (c *Client) defaultHeaders() http.Header {
	h := http.Header{}
	h.Set("User-Agent", defaultUserAgent)
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Origin", "https://www.realtor.ca")
	h.Set("Referer", "https://www.realtor.ca/")
	return h
}
