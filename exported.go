package goreq

import "net/http"

// Get return a get request
func Get(rawURL string) *Req {
	req := New()
	req.rawURL = rawURL
	req.method = http.MethodGet
	return req
}

// Post return a post request
func Post(rawURL string) *Req {
	req := New()
	req.rawURL = rawURL
	req.method = http.MethodPost
	return req
}
