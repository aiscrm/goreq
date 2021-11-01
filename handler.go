package goreq

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io/ioutil"
)

type HandlerFunc func(ctx *Context)
type HandlerChain []HandlerFunc

func DumpHandler() HandlerFunc {
	return func(ctx *Context) {
		ctx.Next()
		ctx.Resp.Request().Body = ioutil.NopCloser(bytes.NewReader(ctx.Req.GetBody()))
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
