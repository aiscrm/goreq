package br

import (
	"bytes"
	"io"

	"github.com/aiscrm/goreq"
	"github.com/andybalholm/brotli"
)

func CompressHandler() goreq.HandlerFunc {
	return func(ctx *goreq.Context) {
		ctx.Next()
		if ctx.Resp.Error() != nil {
			return
		}
		if ctx.Resp.Response().Header.Get(goreq.ContentEncoding) == "br" {
			reader := brotli.NewReader(ctx.Resp.Response().Body)
			body, err := io.ReadAll(reader)
			if err == nil {
				ctx.Resp.Response().Body = io.NopCloser(bytes.NewReader(body))
			}
		}
	}
}
