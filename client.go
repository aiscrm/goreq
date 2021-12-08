package goreq

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

var DefaultClient = NewClient()

type Client interface {
	Init(...Option) error
	Options() Options
	Clone(opts ...Option) Client
	Use(...HandlerFunc) Client
	Do(*Req) *Resp
	New() *Req
	Get(rawURL string) *Req
	Post(rawURL string) *Req
	Put(rawURL string) *Req
	Delete(rawURL string) *Req
	Head(rawURL string) *Req
}

func NewClient(opts ...Option) Client {
	options := NewOptions()
	c := &client{
		options: options,
	}
	_ = c.Init(opts...)
	c.pool.New = func() interface{} {
		return &Context{}
	}

	c.handlers = append(c.handlers, Recovery(), CompressHandler(), c.doHandler())
	return c
}

type client struct {
	options    Options
	httpClient *http.Client
	handlers   HandlerChain
	pool       sync.Pool
}

func (c *client) Init(opts ...Option) error {
	for _, o := range opts {
		o(&c.options)
	}
	// init http client
	c.httpClient = newHTTPClient(c.options)
	return nil
}

func (c *client) Options() Options {
	return c.options
}

func (c *client) Use(handlers ...HandlerFunc) Client {
	finalSize := len(c.handlers) + len(handlers)
	mergedHandlers := make(HandlerChain, finalSize)
	copy(mergedHandlers, c.handlers[:len(c.handlers)-1])
	copy(mergedHandlers[len(c.handlers)-1:finalSize-1], handlers)
	copy(mergedHandlers[finalSize-1:], c.handlers[len(c.handlers)-1:])
	c.handlers = mergedHandlers
	return c
}

func (c *client) Clone(opts ...Option) Client {
	c2 := &client{
		options: c.Options(),
	}
	_ = c2.Init(opts...)
	c2.pool.New = func() interface{} {
		return &Context{}
	}
	c2.handlers = make(HandlerChain, len(c.handlers))
	copy(c2.handlers, c.handlers)
	return c2
}

func (c *client) Do(r *Req) *Resp {
	ctx := c.pool.Get().(*Context)
	ctx.reset()
	defer c.pool.Put(ctx)
	ctx.Req = r
	ctx.Resp = &Resp{}
	ctx.handlers = c.handlers
	ctx.Next()
	return ctx.Resp
}

func (c *client) New() *Req {
	return New().WithClient(c)
}

func (c *client) Get(rawURL string) *Req {
	return c.New().WithURL(rawURL).WithMethod(http.MethodGet)
}

func (c *client) Post(rawURL string) *Req {
	return c.New().WithURL(rawURL).WithMethod(http.MethodPost)
}

func (c *client) Put(rawURL string) *Req {
	return c.New().WithURL(rawURL).WithMethod(http.MethodPut)
}

func (c *client) Delete(rawURL string) *Req {
	return c.New().WithURL(rawURL).WithMethod(http.MethodDelete)
}

func (c *client) Head(rawURL string) *Req {
	return c.New().WithURL(rawURL).WithMethod(http.MethodHead)
}

func newHTTPClient(options Options) *http.Client {
	var jar *cookiejar.Jar
	if options.EnableCookie {
		jar, _ = cookiejar.New(nil)
	}
	transport := options.Transport
	if transport == nil {
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   options.DialTimeout,
				KeepAlive: options.DialKeepAlive,
				// DualStack: true,
			}).DialContext,
			MaxIdleConns:          options.MaxIdleConns,
			IdleConnTimeout:       options.IdleConnTimeout,
			TLSHandshakeTimeout:   options.TLSHandshakeTimeout,
			TLSClientConfig:       options.TLSClientConfig,
			ExpectContinueTimeout: options.ExpectContinueTimeout,
		}
	}
	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   options.Timeout,
	}
}

func (c *client) doHandler() HandlerFunc {
	return func(ctx *Context) {
		if ctx.Req.Error() != nil {
			ctx.Resp.SetError(ctx.Req.Error())
			return
		}
		request, err := ctx.Req.Build()
		if err != nil {
			ctx.Resp.SetError(err)
			return
		}
		ctx.Resp.request = request
		before := time.Now()
		ctx.Resp.response, ctx.Resp.err = c.httpClient.Do(request)
		ctx.Resp.codecs = c.Options().Codecs
		ctx.Resp.cost = time.Since(before)

		if ctx.Resp.err != nil && strings.Contains(ctx.Resp.err.Error(), "Client.Timeout exceeded") { // 超时的判断
			ctx.Resp.timeout = true
		}
	}
}
