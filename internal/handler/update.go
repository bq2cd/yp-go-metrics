package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type updateHandler struct {
	baseHandler
	metrics service.MetricStorer
}

// ServeHTTP implements http.Handler for /update/* endpoint with plain-text requests/responses.
func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metric, err := urlpath.NewOperationFromURLPath(r.URL.Path).ToMetric()
	l := h.logger.With(log.Str("urlpath", r.URL.Path), log.Any("metric", metric))
	switch err {
	case nil:
		switch metric.Type {
		case model.MetricTypeCounter:
			// ok
		case model.MetricTypeGauge:
			// ok
		default:
			l.Error().Msg("unknown metric type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case urlpath.ErrMissingMetricID:
		l.Error().WithErr(err).Msg("missing metric id")
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		l.Error().WithErr(err).Msg("unknown metric operation")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.metrics.StoreSingle(r.Context(), metric)
	if err != nil {
		l.Error().WithErr(err).Msg("cannot store metric")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpheaders.ContentTypeTextPlain.UTF8().Apply(w.Header())
	w.WriteHeader(http.StatusOK)
}
