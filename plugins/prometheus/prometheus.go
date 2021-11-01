package prometheus

import (
	"time"

	"github.com/aiscrm/goreq"

	"github.com/prometheus/client_golang/prometheus"
)

func Prometheus(opts ...Option) goreq.HandlerFunc {
	options := newOptions(opts...)
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: options.NameSpace,
			Name:      "request_total",
			Help:      "Requests processed, partitioned by host, uri and status",
		},
		[]string{
			"host",
			"uri",
			"status",
		},
	)
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  options.NameSpace,
			Name:       "latency_milliseconds",
			Objectives: options.Objectives,
			Help:       "Request latencies in milliseconds, partitioned by host and uri",
		},
		[]string{
			"host",
			"uri",
		},
	)
	options.Registerer.MustRegister(counter)
	options.Registerer.MustRegister(summary)
	return func(ctx *goreq.Context) {
		begin := time.Now()
		ctx.Next()
		d := time.Since(begin)
		summary.WithLabelValues(ctx.Resp.Request().RequestURI).Observe(float64(d.Milliseconds()))
		counter.WithLabelValues(ctx.Resp.Request().URL.Host, ctx.Resp.Request().URL.Path, ctx.Resp.Response().Status).Inc()
	}
	return nil
}
