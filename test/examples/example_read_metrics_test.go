package examples_test

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// This example demonstrates sample usage of `POST /value/` HTTP endpoint.
func Example_readMetrics() {
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

	// Read single metric
	var result Metric

	resp, err = client.R().
		SetBody(Metric{ // Must set `Type` and `ID` fields.
			Type: "counter",
			ID:   "metric-2",
		}).
		SetResult(&result).
		Post("/value/")
	if err != nil {
		panic(err)
	}
	if !resp.IsSuccess() {
		panic(resp.Status())
	}

	// Print obtained metric.
	switch result.Type {
	case "counter":
		fmt.Printf("counter::%s::%d\n", result.ID, *result.Delta)
	case "gauge":
		fmt.Printf("gauge::%s::%.6f\n", result.ID, *result.Value)
	default:
		panic("unsupported type")
	}

	// Output:
	// counter::metric-2::200
}
