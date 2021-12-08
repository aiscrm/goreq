module github.com/aiscrm/goreq/plugins/trace

go 1.17

replace github.com/aiscrm/goreq => ../../

require (
	github.com/aiscrm/goreq v0.2.4
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
)
