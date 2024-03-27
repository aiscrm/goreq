package goreq

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
)

type (
	HandlerFunc  func(*Context)
	HandlerChain []HandlerFunc
)

func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func DumpHandler() HandlerFunc {
	return func(ctx *Context) {
		ctx.Next()
		ctx.Resp.Request().Body = io.NopCloser(bytes.NewReader(ctx.Req.GetBody()))
		fmt.Println(ctx.Resp.Dump())
	}
}

func CompressHandler() HandlerFunc {
	return func(ctx *Context) {
		ctx.Next()
		if ctx.Resp.Error() != nil {
			return
		}
		if ctx.Resp.Response().Header.Get(ContentEncoding) == ContentEncodingGzip {
			body, err := gzip.NewReader(ctx.Resp.Response().Body)
			if err == nil {
				ctx.Resp.Response().Body = body
			}
		}
		if ctx.Resp.Response().Header.Get(ContentEncoding) == ContentEncodingDeflate {
			body, err := zlib.NewReader(ctx.Resp.Response().Body)
			if err != nil {
				ctx.Resp.Response().Body = body
			}
		}
	}
}
