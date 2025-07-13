package goreq

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type (
	HandlerFunc  func(*Context)
	HandlerChain []HandlerFunc
)

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

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

func LogHandler(logger ...Logger) HandlerFunc {
	return func(ctx *Context) {
		ctx.Next()
		ctx.Resp.Request().Body = io.NopCloser(bytes.NewReader(ctx.Req.GetBody()))
		host := ctx.Resp.Request().Host
		method := ctx.Resp.Request().Method
		query := ctx.Resp.Request().URL.Query()
		reqHeaders := ctx.Resp.Request().Header
		reqBody := ctx.Req.GetBody()
		statusCode := ctx.Resp.StatusCode()
		respHeaders := ctx.Resp.Response().Header
		respBody, _ := ctx.Resp.AsBytes()
		var l Logger
		if len(logger) == 0 {
			jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						if t, ok := a.Value.Any().(time.Time); ok {
							a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
						}
					}
					return a
				},
			})
			l = slog.New(jsonHandler)
		} else {
			l = logger[0]
		}
		if statusCode >= http.StatusInternalServerError {
			l.ErrorContext(ctx.Req.Context(), "dump request", "method", method, "host", host, "query", query, "req_headers", reqHeaders, "req_body", string(reqBody), "status_code", statusCode, "resp_headers", respHeaders, "resp_body", string(respBody))
		} else {
			l.InfoContext(ctx.Req.Context(), "dump request", "method", method, "host", host, "query", query, "req_headers", reqHeaders, "req_body", string(reqBody), "status_code", statusCode, "resp_headers", respHeaders, "resp_body", string(respBody))
		}
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
