package codec

type Options struct {
	EscapeHTML   bool
	IndentPrefix string
	IndentValue  string
}

type Option func(*Options)

func WithEscapeHTML(on bool) Option {
	return func(options *Options) {
		options.EscapeHTML = on
	}
}

func WithIndent(prefix, indent string) Option {
	return func(options *Options) {
		options.IndentPrefix = prefix
		options.IndentValue = indent
	}
}
