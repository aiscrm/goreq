package prometheus

import "github.com/prometheus/client_golang/prometheus"

type Options struct {
	NameSpace  string
	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
	Objectives map[float64]float64
}

type Option func(*Options)

func newOptions(opts ...Option) Options {
	options := Options{
		NameSpace:  "goreq",
		Registerer: prometheus.DefaultRegisterer,
		Gatherer:   prometheus.DefaultGatherer,
		Objectives: map[float64]float64{0.0: 0, 0.5: 0.05, 0.75: 0.04, 0.90: 0.03, 0.95: 0.02, 0.98: 0.001, 1: 0},
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func NameSpace(nameSpace string) Option {
	return func(options *Options) {
		options.NameSpace = nameSpace
	}
}

func Registerer(registerer prometheus.Registerer) Option {
	return func(options *Options) {
		options.Registerer = registerer
	}
}

func Gatherer(gatherer prometheus.Gatherer) Option {
	return func(options *Options) {
		options.Gatherer = gatherer
	}
}

func Objectives(objectives map[float64]float64) Option {
	return func(options *Options) {
		options.Objectives = objectives
	}
}
