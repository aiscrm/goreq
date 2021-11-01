package hystrix

import (
	"github.com/aiscrm/goreq"

	"github.com/afex/hystrix-go/hystrix"
)

func Breaker(opts ...Option) goreq.HandlerFunc {
	options := newOptions(opts...)
	return func(ctx *goreq.Context) {
		hystrix.Do(options.KeyFunc(ctx.Req), func() error {
			ctx.Next()
			if ctx.Resp.Timeout() {
				return ctx.Resp.Error()
			}
			return nil
		}, func(err error) error {
			ctx.Resp.SetError(err)
			return nil
		})
	}
}
