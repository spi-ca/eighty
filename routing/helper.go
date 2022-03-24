package routing

import (
	"github.com/spi-ca/eighty"
	"github.com/valyala/fasthttp"
)

// ApplyMiddlware is a function that applies the given middleware for the handler.
func ApplyMiddlware(source Router, middlewares ...Middleware) (handler Router) {
	handler = source
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return
}

// JustCode is a simple handler function that just only returns the http status code.
func JustCode(statusHandler eighty.HandledError, middlewares ...Middleware) (handler Router) {
	return ApplyMiddlware(
		func(ctx *fasthttp.RequestCtx) {
			panic(statusHandler)
		},
		middlewares...,
	)
}
