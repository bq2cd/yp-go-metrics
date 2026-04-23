package examples_test

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// This example demonstrates sample usage of `GET /` HTTP endpoint.
func Example_listMetrics() {
	addr := getRandomLocalhostAddress() // returns "localhost:PORT" or panics

	stop := startDemoServer(addr) // or use `go run ./cmd/server -a ADDR` to launch real server.
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
	resp, err := client.R().
		SetBody(generated).
		Post("/updates/")
	if err != nil {
		panic(err)
	}
	if !resp.IsSuccess() {
		panic(resp.Status())
	}

	// List uploaded metrics
	resp, err = client.R().Get("/")
	if err != nil {
		panic(err)
	}
	if !resp.IsSuccess() {
		panic(resp.Status())
	}

	// Print metrics
	fmt.Println(resp.String())

	// Unordered output:
	// metric-0 0
	// metric-1 0.01
	// metric-2 200
	// metric-3 0.03
	// metric-4 400
}
