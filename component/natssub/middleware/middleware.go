package middleware

import (
	"github.com/nats-io/nats.go/jetstream"
)

type MiddlewareFunc func(handler func(msg jetstream.Msg)) func(msg jetstream.Msg)

func ApplyMiddleware(handler func(msg jetstream.Msg), middlewares ...MiddlewareFunc) func(msg jetstream.Msg) {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
