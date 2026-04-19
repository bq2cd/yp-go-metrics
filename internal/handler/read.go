package handler

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/service"
)

type readHandler struct {
	baseHandler
	metrics service.MetricStorer
}

// ServeHTTP implements [Handler] for /value endpoint
func (h *readHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metrics.RetrieveAll(r.Context())
	if err != nil {
		h.logger.Error().WithErr(err).Msg("cannot retrieve metrics from storage")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpheaders.ContentTypeTextHTML.Apply(w.Header())
	w.WriteHeader(http.StatusOK)

	for _, m := range metrics {
		out := new(bytes.Buffer)
		out.WriteString(m.ID)
		out.WriteByte(' ')
		switch m.Type {
		case model.MetricTypeCounter:
			out.WriteString(strconv.FormatInt(*m.Delta, 10))
		default:
			out.WriteString(strconv.FormatFloat(*m.Value, 'g', 10, 64))
		}
		out.WriteByte('\n')
		_, _ = w.Write(out.Bytes())
	}
}
