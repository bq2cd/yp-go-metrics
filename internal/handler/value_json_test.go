package handler

import (
	"net/http"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/service"
)

func Test_valueJSONHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		metrics service.Metrics
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &valueJSONHandler{
				metrics: tt.fields.metrics,
			}
			h.ServeHTTP(tt.args.w, tt.args.r)
		})
	}
}
