package goreq

import (
	"math"
)

const abortIndex int8 = math.MaxInt8 / 2

type Context struct {
	index    int8
	handlers HandlerChain
	err      error
	Req      *Req
	Resp     *Resp
}

func (c *Context) reset() {
	c.index = -1
	c.Req = nil
	c.Resp = nil
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) AbortWithError(err error) {
	c.err = err
	c.Abort()
}
