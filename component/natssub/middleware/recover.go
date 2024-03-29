package middleware

import (
	"github.com/nats-io/nats.go/jetstream"

	sctx "github.com/phathdt/service-context"
)

func Recover() MiddlewareFunc {
	return func(handler func(msg jetstream.Msg)) func(msg jetstream.Msg) {
		return func(msg jetstream.Msg) {
			defer func() {
				logger := sctx.GlobalLogger().GetLogger("service")

				if r := recover(); r != nil {
					logger.Error("Recovered from panic:", r)
				}
			}()
			handler(msg)
		}
	}
}
