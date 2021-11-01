module github.com/aiscrm/goreq/plugins/breaker/hystrix

go 1.17

replace github.com/aiscrm/goreq => ../../../

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/aiscrm/goreq v0.2.0
)

require github.com/smartystreets/goconvey v1.7.2 // indirect
