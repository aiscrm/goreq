package goreq

// Use handlers for DefaultClient
func Use(handlers ...HandlerFunc) Client {
	DefaultClient.Use(handlers...)
	return DefaultClient
}

// Get return a get request
func Get(rawURL string) *Req {
	return DefaultClient.Get(rawURL)
}

// Post return a post request
func Post(rawURL string) *Req {
	return DefaultClient.Post(rawURL)
}

// Put return a put request
func Put(rawURL string) *Req {
	return DefaultClient.Put(rawURL)
}

// Delete return a delete request
func Delete(rawURL string) *Req {
	return DefaultClient.Delete(rawURL)
}

// Head return a head request
func Head(rawURL string) *Req {
	return DefaultClient.Head(rawURL)
}
