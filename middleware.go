package mux

import "net/http"

// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
// Typically, the returned handler is a closure which does something with the http.ResponseWriter and http.Request passed
// to it, and then calls the handler passed as parameter to the MiddlewareFunc.
type MiddlewareFunc func(http.Handler) http.Handler

// middleware interface is anything which implements a MiddlewareFunc named Middleware.
type middleware interface {
	Middleware(handler http.Handler) http.Handler
}

// MiddlewareFunc also implements the Middleware interface.
func (mw MiddlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}

// AddMiddlewareFunc appends a MiddlewareFunc to the chain.
func (r *Router) AddMiddlewareFunc(mwf MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mwf)
}

// addMiddleware appends a middleware to the chain.
func (r *Router) addMiddleware(mw middleware) {
	r.middlewares = append(r.middlewares, mw)
}
