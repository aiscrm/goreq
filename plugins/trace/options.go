package trace

type Options struct {
	DumpRequest  bool
	DumpResponse bool
	ServiceName  string
}

type Option func(options *Options)

func DumpRequest(dumpRequest bool) Option {
	return func(options *Options) {
		options.DumpRequest = dumpRequest
	}
}

func DumpResponse(dumpResponse bool) Option {
	return func(options *Options) {
		options.DumpResponse = dumpResponse
	}
}

func ServiceName(serviceName string) Option {
	return func(options *Options) {
		options.ServiceName = serviceName
	}
}
