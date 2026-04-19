package examples_test

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// This example demonstrates sample usage of `POST /updates/` HTTP endpoint.
func Example_uploadMetrics() {
	addr := "localhost:12345"

	// Use `go run ./cmd/server` for external testing.
	stop := startDemoServer(addr)
	defer stop()

	client := resty.New().SetBaseURL("http://" + addr)

	type Metric struct {
		ID    string   `json:"id"`
		Type  string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	// Generate some metrics.
	generated := make([]Metric, 5)

	for i := range generated {
		generated[i].ID = fmt.Sprintf("metric-%d", i)
		if i%2 == 0 {
			value := int64(i * 100)
			generated[i].Type = "counter"
			generated[i].Delta = &value
		} else {
			value := float64(i) / float64(100)
			generated[i].Type = "gauge"
			generated[i].Value = &value
		}
	}

	// Upload generated metrics.
	var result []Metric

	resp, err := client.R().
		SetBody(generated).
		SetResult(&result).
		Post("/updates/")
	if err != nil {
		panic(err)
	}
	if !resp.IsSuccess() {
		panic(resp.Status())
	}

	// Print uploaded metrics.
	for _, metric := range result {
		switch metric.Type {
		case "counter":
			fmt.Printf("counter::%s::%d\n", metric.ID, *metric.Delta)
		case "gauge":
			fmt.Printf("gauge::%s::%.6f\n", metric.ID, *metric.Value)
		default:
			panic("unsupported type")
		}
	}

	// Unordered output:
	// counter::metric-0::0
	// gauge::metric-1::0.010000
	// counter::metric-2::200
	// gauge::metric-3::0.030000
	// counter::metric-4::400
}
