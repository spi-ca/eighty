package routing

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"strings"
)

type (
	// RouterRegistry is a fasthttp request url routing builder.
	RouterRegistry interface {
		RouterContext
		// Name returns the current group name.
		Name() string
		// ToContext returns the current group name.
		ToContext() RouterContext
		// Register is a registration method for http request.
		Register(
			name string, path string, params []string,
			handler Router,
			middlewares []Middleware,
			methods ...string,
		)
		// RegisterNested is a registration method for http request with a router generator.
		RegisterNested(
			name string, path string, params []string,
			routerGenerator NestedRouter,
			middlewares []Middleware,
			methods ...string,
		)
		// Wrap returns a child RouterRegistry with specified name and path.
		Wrap(name string, path string, middlewares ...Middleware) RouterRegistry
		// Handler is a handler method that process incoming requests.
		// It implements the Router interface.
		Handler(ctx *fasthttp.RequestCtx)
	}
	routerRegistryImpl struct {
		*routerContextImpl
		name        string
		r           *router.Router
		middlewares []Middleware
		parentNames []string
		parentPaths []string
	}
)

func (r *routerRegistryImpl) Handler(ctx *fasthttp.RequestCtx) {
	r.r.Handler(ctx)
}
func (r *routerRegistryImpl) Name() string             { return r.name }
func (r *routerRegistryImpl) ToContext() RouterContext { return r.routerContextImpl }

func (r *routerRegistryImpl) Register(
	name string, path string, params []string,
	handler Router,
	middlewares []Middleware,
	methods ...string,
) {

	fullPath := r.reverseRouter.MustAddGr(name, path, r.parentNames, r.parentPaths, params...)
	mixedRouter := ApplyMiddlware(handler, append(r.middlewares, middlewares...)...)
	for _, method := range methods {
		r.r.Handle(method, fullPath, mixedRouter)
	}
}

func (r *routerRegistryImpl) RegisterNested(
	name string, path string, params []string,
	routerGenerator NestedRouter,
	middlewares []Middleware,
	methods ...string,
) {
	r.Register(name, path, params, routerGenerator(), middlewares, methods...)
}

func (r *routerRegistryImpl) Wrap(name string, path string, middlewares ...Middleware) RouterRegistry {
	var (
		newName, newPath []string
	)
	if len(name) > 0 {
		newName = append(r.parentNames, name)
	} else {
		newName = r.parentNames
	}
	if len(path) > 0 {
		newPath = append(r.parentPaths, path)
	} else {
		newPath = r.parentPaths
	}
	return &routerRegistryImpl{
		routerContextImpl: r.routerContextImpl,
		r:                 r.r,
		middlewares:       append(r.middlewares, middlewares...),
		name:              strings.Join(newName, "."),
		parentNames:       newName,
		parentPaths:       newPath,
	}
}
func (r *routerRegistryImpl) String() string { return r.urlFor().String() }
