package routing

import (
	"github.com/fasthttp/router"
	"net/http"
)

type (
	// RouterContext is a fasthttp request url routing context.
	RouterContext interface {
		// UrlResolver returns a UrlResolver.
		UrlResolver() UrlResolver
		urlFor() UrlFor
		// UrlPrefix returns the URL prefix that is added across the entire routing group.
		UrlPrefix() string
		withRegister(
			r *router.Router,
			parentNames []string,
			parentPaths []string,
			middlewares ...Middleware,
		) RouterRegistry
		// BuildRouter returns a RouterRegistry for register grouped routing.
		BuildRouter(errHandleMiddleware Middleware) RouterRegistry
	}

	routerContextImpl struct {
		urlPrefix     string
		reverseRouter UrlFor
	}
)

func (ctx *routerContextImpl) UrlResolver() UrlResolver { return ctx.reverseRouter.ToResolver() }
func (ctx *routerContextImpl) urlFor() UrlFor           { return ctx.reverseRouter }
func (ctx *routerContextImpl) UrlPrefix() string        { return ctx.urlPrefix }
func (ctx *routerContextImpl) withRegister(
	r *router.Router,
	parentNames []string,
	parentPaths []string,
	middlewares ...Middleware,
) RouterRegistry {
	return &routerRegistryImpl{
		routerContextImpl: ctx,
		r:                 r,
		middlewares:       middlewares,
		parentNames:       parentNames,
		parentPaths:       parentPaths,
	}
}
func (c *routerContextImpl) BuildRouter(errHandlemiddleware Middleware) RouterRegistry {
	r := router.New()
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true
	r.HandleOPTIONS = false
	r.HandleMethodNotAllowed = true
	if errHandlemiddleware != nil {
		r.NotFound = JustCode(http.StatusNotFound, errHandlemiddleware)
		r.MethodNotAllowed = JustCode(http.StatusMethodNotAllowed, errHandlemiddleware)
	}
	return &routerRegistryImpl{routerContextImpl: c, r: r}
}

// NewRouterContext returns a RouterContext.
func NewRouterContext(
	urlPrefix string,
	reverseRouter UrlFor,
) (ctx RouterContext) {
	return &routerContextImpl{
		urlPrefix:     urlPrefix,
		reverseRouter: reverseRouter,
	}
}
