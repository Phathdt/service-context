package sctx

import (
	"flag"
)

var (
	EnableTracing = false
)

func init() {
	flag.BoolVar(&EnableTracing, "enable-tracing", false, "enable tracing")
	flag.Parse()
}
