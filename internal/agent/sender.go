package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
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
	Send(metric model.Metric) error
}

// NewSenderPlain creates an instance of a sender that reports metrics in plain text.
func NewSenderPlain(ctx context.Context, client *resty.Client) *senderPlain {
	return &senderPlain{context: ctx, client: client}
}

type senderPlain struct {
	context context.Context
	client  *resty.Client
}

func (s *senderPlain) Send(metric model.Metric) error {
	metricOp := urlpath.NewOperationFromMetric(urlpath.OperationTypeUpdate, metric)
	urlPath, err := metricOp.ToURLPath()
	if err != nil {
		return fmt.Errorf("cannot convert metric to url path: %w", err)
	}

	req := s.client.R().SetContext(s.context)

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
func NewSenderJSON(ctx context.Context, client *resty.Client) *senderJSON {
	return &senderJSON{context: ctx, client: client, shouldCompress: true}
}

type senderJSON struct {
	context        context.Context
	client         *resty.Client
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

func (s *senderJSON) Send(metric model.Metric) error {
	if metric.Empty() {
		return ErrSenderEmptyMetric
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		return fmt.Errorf("json encoder failed: %w", err)
	}

	req := s.client.R().SetContext(s.context).SetHeader(httpheaders.HeaderKeyContentType, httpheaders.ContentTypeApplicationJSON.String())

	if err := s.setBody(req, &buf); err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	resp, err := req.Post("/update/")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSenderRequestFailed, err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("expected success, got status %v: %w", resp.Status(), ErrSenderResponseNotOK)
	}

	return nil
}
