// Package mux wraps the chi router to expose all HTTP verbs while adding support
// for per-route scope annotations.
//
// Routes may carry a "[scope:...]" annotation in their pattern, which mux strips
// before registration and records in a route tree so scopes can be looked up
// during a request and enforced by middleware.
package mux

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Mux package is a wrapper designed to work with Chi. The purpose is two fold, expose all
// HTTP verbs that one will need to run a full server as well as setup a simple approach
// to introducing new methods on top of what the base package provides. We are drafting
// this package to help us introduce a cache to store the route and scope in memory and
// expose a method to check that a route's scopes during an HTTP request using a middleware.
// The usage is exactly the same as what we'd need to run a Chi router. But, we are going
// to call our wrapper to get things kicked off. For example, create a new router:
// a := NewRouter()
//
//	... now you are able to call any Chi method:
//
// a.use()
func NewRouter() *Mux {
	return &Mux{chi.NewRouter()}
}

func (r *Mux) URLParam(rq *http.Request, key string) string {
	return chi.URLParam(rq, key)
}

// Mux route tree traversal to locate a route and return the scopes attached
// to the pattern.
func (r *Mux) GetScopes(path string) MuxRouteScope {
	var scope MuxRouteScope
	for _, router := range MuxRouterTree {
		if router.Base+router.Route == path {
			if strings.TrimSpace(router.Scope) != "" {
				scope.Scope = strings.Split(router.Scope, " ")
			}
			break
		}
	}
	return scope
}

// With adds inline middlewares for an endpoint handler.
func (r *Mux) With(middlewares ...func(http.Handler) http.Handler) chi.Router {
	mx := r.Mux.With(middlewares...).(*chi.Mux)
	return mx
}

// Use appends a middleware handler to the Mux middleware stack.
// The middleware stack for any Mux will execute before searching for a
// matching route to a specific handler, which provides opportunity to respond
// early, change the course of the request execution, or set request-scoped values
// for the next http.Handler.
func (r *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	r.Mux.Use(middlewares...)
}

// Handle adds the route `pattern` that matches any http method to execute the
// `handler` http.Handler.
func (r *Mux) Handle(pattern string, handler http.Handler) {
	r.Mux.Handle(cleanMuxScopeAnnotation(pattern), handler)
}

// HandleFunc adds the route `pattern` that matches any http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.Mux.HandleFunc(cleanMuxScopeAnnotation(pattern), handler)
}

// Match searches the routing tree for a handler that matches the method/path. It's
// similar to routing a http request, but without executing the handler thereafter.
// Note: the *Context state is updated during execution, so manage the state carefully
// or make a NewRouteContext().
func (r *Mux) Match(rctx *chi.Context, method string, path string) bool {
	return r.Mux.Match(rctx, method, path)
}

// Method and MethodFunc adds routes for `pattern` that matches the `method` HTTP method.
func (r *Mux) Method(method, pattern string, handler http.Handler) {
	r.Mux.With().Method(method, cleanMuxScopeAnnotation(pattern), handler)
}

// Method and MethodFunc adds routes for `pattern` that matches
// the `method` HTTP method.
func (r *Mux) MethodFunc(method, pattern string, handler http.HandlerFunc) {
	r.Mux.With().MethodFunc(method, cleanMuxScopeAnnotation(pattern), handler)
}

// Connect adds the route `pattern` that matches a CONNECT http method to execute
// the `handlerFn` http.HandlerFunc.
func (r *Mux) Connect(pattern string, handler http.HandlerFunc) {
	r.Mux.Connect(cleanMuxScopeAnnotation(pattern), handler)
}

// Find searches the routing tree for the pattern that matches
// the method/path.
func (r *Mux) Find(rctx *chi.Context, method, path string) string {
	return r.Mux.Find(rctx, method, path)
}

// Head adds the route `pattern` that matches a HEAD http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Head(pattern string, handler http.HandlerFunc) {
	r.Mux.Head(cleanMuxScopeAnnotation(pattern), handler)
}

// Get adds the route `pattern` that matches a GET http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Get(pattern string, handler http.HandlerFunc) {
	r.Mux.Get(cleanMuxScopeAnnotation(pattern), handler)
}

// Post adds the route `pattern` that matches a POST http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Post(pattern string, handler http.HandlerFunc) {
	r.Mux.Post(cleanMuxScopeAnnotation(pattern), handler)
}

// Put adds the route `pattern` that matches a PUT http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Put(pattern string, handler http.HandlerFunc) {
	r.Mux.Put(cleanMuxScopeAnnotation(pattern), handler)
}

// Patch adds the route `pattern` that matches a PATCH http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Patch(pattern string, handler http.HandlerFunc) {
	r.Mux.Patch(cleanMuxScopeAnnotation(pattern), handler)
}

// Delete adds the route `pattern` that matches a DELETE http method to execute
// the `handlerFn` http.HandlerFunc.
func (r *Mux) Delete(pattern string, handler http.HandlerFunc) {
	r.Mux.Delete(cleanMuxScopeAnnotation(pattern), handler)
}

// Trace adds the route `pattern` that matches a TRACE http method to execute the
// `handlerFn` http.HandlerFunc.
func (r *Mux) Trace(pattern string, handler http.HandlerFunc) {
	r.Mux.Trace(cleanMuxScopeAnnotation(pattern), handler)
}

// Options adds the route `pattern` that matches an OPTIONS http method to execute
// the `handlerFn` http.HandlerFunc.
func (r *Mux) Options(pattern string, handler http.HandlerFunc) {
	r.Mux.Options(cleanMuxScopeAnnotation(pattern), handler)
}

// NotFound sets a custom http.HandlerFunc for routing paths that could not
//
//	be found. The default 404 handler is `http.NotFound`.
func (r *Mux) NotFound(handlerFn http.HandlerFunc) {
	r.Mux.NotFound(handlerFn)
}

// NotFoundHandler returns the default Mux 404 responder whenever a route cannot
// be found.
func (r *Mux) NotFoundHandler() http.HandlerFunc {
	return r.Mux.NotFoundHandler()
}

// MethodNotAllowed sets a custom http.HandlerFunc for routing paths where the
// method is unresolved. The default handler returns a 405 with an empty body.
func (r *Mux) MethodNotAllowed(handlerFn http.HandlerFunc) {
	r.Mux.MethodNotAllowed(handlerFn)
}

// ServeHTTP is the single method of the http.Handler interface that makes Mux
// interoperable with the standard library. It uses a sync.Pool to get and reuse
// routing contexts for each request.
func (r *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}

// Group creates a new inline-Mux with a copy of middleware stack. It's useful
//
//	for a group of handlers along the same routing path that use an additional
//
// set of middlewares.
func (r *Mux) Group(fn func(r chi.Router)) chi.Router {
	return r.Mux.Group(fn)
}

// Route creates a new Mux and mounts it along the `pattern` as a subrouter.
// Effectively, this is a short-hand call to Mount.
func (r *Mux) Route(pattern string, fn func(r chi.Router)) chi.Router {
	return r.Mux.Route(pattern, fn)
}

// Mount attaches another http.Handler or chi Router as a subrouter along a routing
// path. It's very useful to split up a large API as many independent routers and
// compose them as a single service using Mount.
func (r *Mux) Mount(pattern string, handler http.Handler) {

	r.Mux.Mount(pattern, handler)

	var extractRoute func([]chi.Route, string)
	processRouter := func(r chi.Router, basePattern string) {
		extractRoute(r.Routes(), basePattern)
	}

	extractRoute = func(routes []chi.Route, basePattern string) {
		for _, route := range routes {

			// range the mux tree for the matching pattern
			for i, muxRoute := range MuxRouterTree {
				if muxRoute.Route == route.Pattern {
					muxRoute.Base = basePattern
					muxRoute.Scope = extractScopeFromMuxPattern(muxRoute.Annotation)
					MuxRouterTree[i] = muxRoute
					break
				}
			}

			// check if the handler is another chi router aka subrouter
			if reflect.TypeOf(route.SubRoutes) != nil {
				extractRoute(route.SubRoutes.Routes(), basePattern)
			}
		}
	}

	processRouter(r, pattern)
}

// Middlewares returns a slice of middleware handler functions.
func (r *Mux) Middlewares() chi.Middlewares {
	return r.Mux.Middlewares()
}

// Routes returns a slice of routing information from the tree, useful
// for traversing available routes of a router.
func (r *Mux) Routes() []chi.Route {
	return r.Mux.Routes()
}

// Clean the mux pattern, capture the values in the mux node tree and
// return the pattern used for HTTP routing. The scope annotation i.e.,
// string pattern is enclosed in square brackets.
func cleanMuxScopeAnnotation(pattern string) string {

	re := regexp.MustCompile(`(?:\[|\])`)
	hasAnnotation := re.MatchString(pattern)

	// nothing to do here if the pattern has no annotation
	if !hasAnnotation {
		MuxRouterTree = append(MuxRouterTree, MuxRouteInfo{
			Route: pattern,
		})
		return pattern
	}

	re = regexp.MustCompile(`(.*)(?:\[(.*)\])`)
	annotation := re.FindStringSubmatch(pattern)
	if len(annotation) != 3 {
		panic("adele: detected malformed annotation in pattern; " + pattern)
	}

	MuxRouterTree = append(MuxRouterTree, MuxRouteInfo{
		Annotation: pattern,
		Route:      annotation[1],
	})

	return annotation[1]
}

// Extract a "scope" value from a given string pattern, where the scope
// is defined within square brackets [] and follows a specific format
// (e.g., [scope:value] or [scopes:value]). The extracted scopes, scopes
// assigned to the route, are returned.
func extractScopeFromMuxPattern(pattern string) string {
	re := regexp.MustCompile(`(?:\[|\])`)
	hasAnnotation := re.MatchString(pattern)
	if !hasAnnotation {
		return ""
	}

	re = regexp.MustCompile(`(.*)(?:\[(.*)\])`)
	annotation := re.FindStringSubmatch(pattern)
	if len(annotation) != 3 {
		panic("adele: detected malformed annotation in pattern; " + pattern)
	}

	typ, val, has := strings.Cut(annotation[2], ":")
	if !has {
		panic("adele: detected malformed annotation type in pattern: " + annotation[2])
	}

	typ = strings.TrimSpace(typ)

	if typ == "scopes" || typ == "scope" {
		return strings.TrimSpace(val)
	}

	panic("adele: detected unknown annotation type in pattern: " + typ)

}
