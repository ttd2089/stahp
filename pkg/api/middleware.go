package api

import (
	"net/http"
)

// A Middleware wraps an [http.Handler] to provide additional processing before and/or after the
// wrapped handler executes, and possibly handle the request directly without invoking the wrapped
// handler at all.
type Middleware interface {

	// Wrap returns a new [http.Handler] that performs additional processing before and/or after
	// forwarding the request to the wrapped handler. Depending on the purpose of the middleware,
	// some requests may not be forwarded to the wrapped handler at all.
	Wrap(http.Handler) http.Handler
}

// A MiddlewareFunc is a function that implements [Middleware].
type MiddlewareFunc func(http.Handler) http.Handler

// Wrap implements [Middleware].
func (fn MiddlewareFunc) Wrap(next http.Handler) http.Handler {
	return fn(next)
}
