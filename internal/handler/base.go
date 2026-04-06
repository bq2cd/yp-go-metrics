package handler

import (
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

type baseHandler struct {
	logger log.Logger
}

func (h *baseHandler) setLogger(logger log.Logger) {
	h.logger = logger
}

func (h *baseHandler) getLogger() log.Logger {
	return h.logger
}

func (h *baseHandler) respondError(w http.ResponseWriter, status int, l log.Logger, err error, msg string) {
	ev := l.Error()
	if err != nil {
		ev = ev.WithErr(err)
	}

	ev.Msg(msg)
	w.WriteHeader(status)
}

func (h *baseHandler) getClientInfo(r *http.Request) model.ClientInfo {
	return model.ClientInfo{}
}
