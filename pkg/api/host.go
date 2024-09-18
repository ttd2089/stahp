package api

import (
	"context"
	"net/http"
)

type Host struct {
	httpServer *http.Server
}

func (h *Host) Run() error {
	return h.httpServer.ListenAndServe()
}

func (h *Host) Stop(ctx context.Context) error {
	return h.httpServer.Shutdown(ctx)
}
