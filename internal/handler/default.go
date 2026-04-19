package handler

import (
	"net/http"
)

type defaultHandler struct {
	baseHandler
}

// ServeHTTP implements [Handler] and serves all unknown endpoints.
func (h *defaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
