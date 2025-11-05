package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
)

var (
	ErrSenderRequestFailed = errors.New("http request error")
	ErrSenderResponseNotOK = errors.New("http response not OK")
	ErrSenderEmptyMetric   = errors.New("empty metric")
)

// Sender abstracts a protocol to encode and send a single metric.
type Sender interface {
	Send(ctx context.Context, metric model.Metric) error
}

// SendBatch allows to send metrics in batches.
type SenderBatch interface {
	Sender
	SendBatch(ctx context.Context, metrics model.MetricSet) (model.MetricSet, error)
}

// NewSenderPlain creates an instance of a sender that reports metrics in plain text.
func NewSenderPlain(client *resty.Client) *senderPlain {
	return &senderPlain{client: client}
}

type senderPlain struct {
	client *resty.Client
}

func (s *senderPlain) Send(ctx context.Context, metric model.Metric) error {
	metricOp := urlpath.NewOperationFromMetric(urlpath.OperationTypeUpdate, metric)
	urlPath, err := metricOp.ToURLPath()
	if err != nil {
		return fmt.Errorf("cannot convert metric to url path: %w", err)
	}

	req := s.client.R().SetContext(ctx)

	resp, err := req.SetHeader("content-type", "text/plain").Post(urlPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSenderRequestFailed, err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected success, got status %v: %w", resp.Status(), ErrSenderResponseNotOK)
	}

	return nil
}

// NewSenderJSON creates an instance of a sender that reports metrics encoded in JSON.
func NewSenderJSON(client *resty.Client, retrierFactory retrymgr.RetrierFactory, hmacSigner hmacsigner.HMACSigner) *senderJSON {
	return &senderJSON{
		client:         client,
		retrierFactory: retrierFactory,
		hmacSigner:     hmacSigner,
		shouldCompress: true,
	}
}

type senderJSON struct {
	client         *resty.Client
	retrierFactory retrymgr.RetrierFactory
	hmacSigner     hmacsigner.HMACSigner
	shouldCompress bool
}

func (s *senderJSON) setBody(req *resty.Request, r io.Reader) error {
	if !s.shouldCompress {
		req.SetBody(r)
		return nil
	}
	var buf bytes.Buffer
	wgz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return err
	}
	_, err = io.Copy(wgz, r)
	if err != nil {
		return err
	}
	err = wgz.Close()
	if err != nil {
		return err
	}
	req.SetHeader(httpheaders.HeaderKeyContentEncoding, httpheaders.ContentEncodingGzip.String())
	req.SetBody(&buf)
	return nil
}

func (s *senderJSON) sendSingleRequest(ctx context.Context, method, url string, body []byte) (*resty.Response, error) {
	req := s.client.R().
		SetContext(ctx).
		SetHeader(httpheaders.HeaderKeyContentType, httpheaders.ContentTypeApplicationJSON.String())

	req.Method = method
	req.URL = url

	signature, err := s.hmacSigner.Sign(body)
	switch {
	case errors.Is(err, hmacsigner.ErrMissingSecretKey):
		// no secret key, no need to add any headers
	case err == nil:
		hash := httpheaders.GetHashSHA256FromBytes(signature)
		req.SetHeader(httpheaders.HeaderKeyHashSHA256, hash.String())
	default:
		return nil, fmt.Errorf("cannot sign request body: %w", err)
	}

	err = s.setBody(req, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot compress request body: %w", err)
	}
	resp, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSenderRequestFailed, err)
	}
	if !resp.IsSuccess() {
		return resp, fmt.Errorf("expected success, got status %v: %w", resp.Status(), ErrSenderResponseNotOK)
	}
	return resp, nil
}

func (s *senderJSON) sendWithRetries(ctx context.Context, method, url string, body []byte) (*resty.Response, error) {
	return retrymgr.NewRetrier[*resty.Response](s.retrierFactory).Do(
		ctx, "send_metrics_json",
		func(ctx context.Context) (*resty.Response, error) {
			return s.sendSingleRequest(ctx, method, url, body)
		},
		func(err error) bool {
			if errors.Is(err, ErrSenderRequestFailed) {
				return true
			}
			if errors.Is(err, ErrSenderResponseNotOK) {
				return true
			}
			return false
		},
	)
}

func (s *senderJSON) Send(ctx context.Context, metric model.Metric) error {
	if metric.Empty() {
		return ErrSenderEmptyMetric
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		return fmt.Errorf("json encoder failed: %w", err)
	}

	_, err := s.sendWithRetries(ctx, http.MethodPost, "/update/", buf.Bytes())
	return err
}

func (s *senderJSON) SendBatch(ctx context.Context, metrics model.MetricSet) (model.MetricSet, error) {
	if metrics.Empty() {
		return model.NewMetricSet(), nil
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metrics.Values()); err != nil {
		return model.NewMetricSet(), fmt.Errorf("json encoder failed: %w", err)
	}

	resp, err := s.sendWithRetries(ctx, http.MethodPost, "/updates/", buf.Bytes())
	if err != nil {
		return model.NewMetricSet(), err
	}

	updated := make([]model.Metric, 0, len(metrics))
	err = json.Unmarshal(resp.Body(), &updated)
	if err != nil {
		return model.NewMetricSet(), fmt.Errorf("cannot unmarshal response (%v): %w", string(resp.Body()), err)
	}

	return model.NewMetricSet(updated...), nil
}
