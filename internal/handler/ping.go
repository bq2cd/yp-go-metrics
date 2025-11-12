package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type pingHandler struct {
	baseHandler
	pinger  service.StoragePinger
	timeout time.Duration
}

// ServeHTTP implements [http.Handler] for /ping endpoint.
func (h *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.pinger.Ping(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Error().WithErr(err).Msg("database unreachable")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case <-ctx.Done():
		h.logger.Error().Dur("timeout", h.timeout).Msg("database timed out")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpheaders.ContentTypeTextPlain.Apply(w.Header())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`OK`))
}
