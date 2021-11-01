package hystrix

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/aiscrm/goreq"
)

type CommandConfig struct {
	Timeout                int `json:"timeout"`                  // how long to wait for command to complete, in milliseconds, default 1000
	MaxConcurrentRequests  int `json:"max_concurrent_requests"`  // how many commands of the same type can run at the same time, default 10
	RequestVolumeThreshold int `json:"request_volume_threshold"` // the minimum number of requests needed before a circuit can be tripped due to health, default 20
	SleepWindow            int `json:"sleep_window"`             // how long, in milliseconds, to wait after a circuit opens before testing for recovery, default 5000
	ErrorPercentThreshold  int `json:"error_percent_threshold"`  // causes circuits to open once the rolling measure of errors exceeds this percent of requests, default 50
}

type Options struct {
	KeyFunc  func(request *goreq.Req) string
	Commands map[string]hystrix.CommandConfig
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		KeyFunc: func(r *goreq.Req) string {
			return r.GetURL()
		},
		Commands: make(map[string]hystrix.CommandConfig),
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func KeyFunc(keyFunc func(request *goreq.Req) string) Option {
	return func(options *Options) {
		options.KeyFunc = keyFunc
	}
}

func ConfigureCommand(name string, config CommandConfig) Option {
	return func(options *Options) {
		commandConfig := hystrix.CommandConfig{
			Timeout:                config.Timeout,
			MaxConcurrentRequests:  config.MaxConcurrentRequests,
			RequestVolumeThreshold: config.RequestVolumeThreshold,
			SleepWindow:            config.SleepWindow,
			ErrorPercentThreshold:  config.ErrorPercentThreshold,
		}
		options.Commands[name] = commandConfig
		hystrix.ConfigureCommand(name, commandConfig)
	}
}

func DefaultTimeout(timeout int) Option {
	return func(*Options) {
		hystrix.DefaultTimeout = timeout
	}
}

func DefaultMaxConcurrentRequests(maxConcurrentRequests int) Option {
	return func(*Options) {
		hystrix.DefaultMaxConcurrent = maxConcurrentRequests
	}
}
func DefaultRequestVolumeThreshold(requestVolumeThreshold int) Option {
	return func(*Options) {
		hystrix.DefaultVolumeThreshold = requestVolumeThreshold
	}
}
func DefaultSleepWindow(sleepWindow int) Option {
	return func(*Options) {
		hystrix.DefaultSleepWindow = sleepWindow
	}
}
func DefaultErrorPercentThreshold(errorPercentThreshold int) Option {
	return func(*Options) {
		hystrix.DefaultErrorPercentThreshold = errorPercentThreshold
	}
}
