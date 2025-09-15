package handler

import (
	"net/http"
)

type defaultHandler struct{}

// ServeHTTP implements http.Handler for any endpoint except /update
func (h *defaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		http.Error(w, "", http.StatusBadRequest)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}
