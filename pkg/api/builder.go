package api

import "net/http"

type Builder struct {
	routes map[string]Endpointer
}

func (b *Builder) AddEndpoint(path string, endpoint Endpointer) {
	if b.routes == nil {
		b.routes = make(map[string]Endpointer)
	}
	b.routes[path] = endpoint
}

func (b *Builder) Build(resolver Resolver) (Host, error) {

	mux := new(http.ServeMux)

	for path, ep := range b.routes {
		mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scope := resolver.NewScope()
			endpoint := ep.Endpoint(scope)
			endpoint.ServeHTTP(w, r)
		}))
	}

	srv := new(http.Server)
	srv.Handler = mux

	return Host{
		httpServer: srv,
	}, nil
}
