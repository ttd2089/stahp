package api

import "net/http"

type Endpointer interface {
	Endpoint(Resolver) http.Handler
}
