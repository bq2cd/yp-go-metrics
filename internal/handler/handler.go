package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

const (
	_contentTypeHeaderKey      = "Content-Type"
	_contentTypeEmpty          = contentType("")
	contentTypeTextPlain       = contentType("text/plain")
	contentTypeTextPlainUTF8   = contentType("text/plain; charset=utf-8")
	contentTypeApplicationJSON = contentType("application/json")
)

type contentType string

func (c contentType) applyToRequest(r *http.Request) {
	if c == _contentTypeEmpty {
		return
	}
	r.Header.Set(_contentTypeHeaderKey, string(c))
}

func (c contentType) applyToResponse(w http.ResponseWriter) {
	if c == _contentTypeEmpty {
		return
	}
	w.Header().Set(_contentTypeHeaderKey, string(c))
}

func (c contentType) matchesRequest(r *http.Request) bool {
	target := contentType(r.Header.Get(_contentTypeHeaderKey))
	return target == c
}

type metricJSONResponder interface {
	WriteResponse(w http.ResponseWriter, m model.Metric) error
}

type defaultMetricJSONResponder struct{}

func (r *defaultMetricJSONResponder) WriteResponse(w http.ResponseWriter, m model.Metric) error {
	contentTypeApplicationJSON.applyToResponse(w)
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(m)
}
