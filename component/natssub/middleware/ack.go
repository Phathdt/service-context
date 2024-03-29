package middleware

import "github.com/nats-io/nats.go/jetstream"

func Ack() MiddlewareFunc {
	return func(handler func(msg jetstream.Msg)) func(msg jetstream.Msg) {
		return func(msg jetstream.Msg) {
			defer msg.Ack()

			handler(msg)
		}
	}
}
