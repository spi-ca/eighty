package routing

import (
	"github.com/valyala/fasthttp"
)

type (
	// Router is an alias type of the fasthttp.RequestHandler.
	Router = fasthttp.RequestHandler
	// NestedRouter is a generator function that returns the Router handler.
	NestedRouter func() Router
	// Middleware is a wrapping function that filters the request.
	Middleware = func(next Router) Router
)
