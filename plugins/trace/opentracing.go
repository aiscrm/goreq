package trace

import (
	"net/http/httputil"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/aiscrm/goreq"
)

func Trace(opts ...Option) goreq.HandlerFunc {
	options := Options{}
	for _, o := range opts {
		o(&options)
	}
	tracer := otel.Tracer(options.ServiceName)
	return func(ctx *goreq.Context) {
		_, span := tracer.Start(ctx.Req.Context(), getSpanName(ctx.Req.GetMethod(), ctx.Req.GetURL()))
		defer span.End()
		ctx.Next()
		span.SetAttributes(attribute.Key("http.method").String(ctx.Req.GetMethod()))
		span.SetAttributes(attribute.Key("http.status_code").Int(ctx.Resp.StatusCode()))
		span.SetAttributes(attribute.Key("http.url").String(ctx.Resp.Request().URL.RequestURI()))
		if ctx.Resp.Error() != nil {
			span.RecordError(ctx.Resp.Error())
			span.SetStatus(codes.Error, ctx.Resp.Error().Error())
		}
		if options.DumpRequest {
			reqData, _ := httputil.DumpRequest(ctx.Resp.Request(), true)
			span.AddEvent("dump.request", trace.WithAttributes(attribute.Key("body").String(string(reqData))))
		}
		if options.DumpResponse {
			respData, _ := httputil.DumpResponse(ctx.Resp.Response(), true)
			span.AddEvent("dump.response", trace.WithAttributes(attribute.Key("body").String(string(respData))))
		}
	}
}

func getSpanName(str ...string) string {
	sb := strings.Builder{}
	sb.WriteString("req")
	for _, s := range str {
		sb.WriteString(":")
		sb.WriteString(s)
	}
	return sb.String()
}
