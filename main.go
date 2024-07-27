// Package stahp provides enablement mechanisms for marshaling HTTP requests to strongly-typed
// handlers functions and marshaling the results back as a response.
package stahp

import (
	"context"
	"net/http"
)

// A Target is a strongly-typed function taking a request and returning a response or an error.
// Marshaling instances of HTTP requests to a [Target] and marshaling the response or the error
// back as a response is the mission of the [stahp] package.
type Target[Req any, Resp any] func(context.Context, Req) (Resp, error)

// A NoReqTarget is a function that can be used with [NoReq] to create a [Target] that doesn't need
// any input from the HTTP request.
type NoReqTarget[Resp any] func(context.Context) (Resp, error)

// NoReq builds a [Target] from a [NoReqTarget].
func NoReq[Resp any](target NoReqTarget[Resp]) Target[struct{}, Resp] {
	return func(ctx context.Context, _ struct{}) (Resp, error) {
		return target(ctx)
	}
}

// A RequestParser extracts a strongly-typed request from an HTTP request.
type RequestParser[Req any] func(*http.Request) (Req, error)

// NoReqParser is a convenience function to satisfy the requirement for a [RequestParser] when
// marshaling requests to a [NoReqParser].
func NoReqParser(*http.Request) (struct{}, error) {
	return struct{}{}, nil
}

// A ResponseWriter marshals a value to an HTTP response.
type ResponseWriter[T any] func(http.ResponseWriter, T)

// A Responder marshals responses and errors to an HTTP response.
type Responder[Resp any] interface {

	// Write marshals responses from a [Target] to an HTTP response.
	Write(http.ResponseWriter, Resp)

	// WriteParseErr marshals errors that occur parsing an HTTP request.
	WriteParseErr(http.ResponseWriter, error)

	// WriteErr marshals errors from [Target] to an HTTP response.
	WriteErr(http.ResponseWriter, error)
}

// NewResponder builds a [Responder] from a [ResponseWriter] for each
func NewResponder[Resp any](
	write ResponseWriter[Resp],
	writeParseErr ResponseWriter[error],
	writeErr ResponseWriter[error],
) Responder[Resp] {
	return responder[Resp]{
		write,
		writeParseErr,
		writeErr,
	}
}

type responder[Resp any] struct {
	write         ResponseWriter[Resp]
	writeParseErr ResponseWriter[error]
	writeErr      ResponseWriter[error]
}

func (r responder[Resp]) Write(w http.ResponseWriter, resp Resp) {
	r.write(w, resp)
}

func (r responder[Resp]) WriteParseErr(w http.ResponseWriter, err error) {
	r.writeParseErr(w, err)
}

func (r responder[Resp]) WriteErr(w http.ResponseWriter, err error) {
	r.writeErr(w, err)
}

// Route generates an [http.HandlerFunc] from a [RequestParser], a [Target], and a [Responder].
func Route[Req any, Resp any](
	target Target[Req, Resp],
	parser RequestParser[Req],
	responder Responder[Resp],
) http.Handler {
	return route[Req, Resp]{
		target,
		parser,
		responder,
	}
}

type route[Req any, Resp any] struct {
	target    Target[Req, Resp]
	parse     RequestParser[Req]
	responder Responder[Resp]
}

func (r route[Req, Resp]) ServeHTTP(w http.ResponseWriter, rr *http.Request) {
	req, err := r.parse(rr)
	if err != nil {
		r.responder.WriteParseErr(w, err)
		return
	}
	resp, err := r.target(rr.Context(), req)
	if err != nil {
		r.responder.WriteErr(w, err)
		return
	}
	r.responder.Write(w, resp)
}
