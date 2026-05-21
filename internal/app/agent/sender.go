package agent

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/handler/urlpath"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/bufpool"
	"github.com/bq2cd/yp-go-metrics/pkg/gzippool"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
)

var (
	// ErrSenderRequestFailed is returned by [Sender.Send] or [SenderBatch.SendBatch] when
	// HTTP request to an upstream server fails due to network errors.
	ErrSenderRequestFailed = errors.New("http request error")
	// ErrSenderResponseNotOK is returned by [Sender.Send] or [SenderBatch.SendBatch] when
	// HTTP response from an upstream server has non-200 status.
	ErrSenderResponseNotOK = errors.New("http response not OK")
	// ErrSenderEmptyMetric is returned by [Sender.Send] when attempting to send a single metric
	// without value (such metric would not be accepted by the server).
	ErrSenderEmptyMetric = errors.New("empty metric")
)

// Sender abstracts a protocol to encode and send a single metric.
type Sender interface {
	Send(ctx context.Context, metric model.Metric) error
}

// SenderBatch allows to send metrics in batches.
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
func NewSenderJSON(
	client *resty.Client,
	retrierFactory retrymgr.RetrierFactory,
	hmacSigner hmacsigner.HMACSigner,
	encryptor asymcrypt.Encryptor,
) *senderJSON {
	compressorPool, _ := gzippool.NewWriterPool(gzip.BestSpeed)

	return &senderJSON{
		client:         client,
		retrierFactory: retrierFactory,
		hmacSigner:     hmacSigner,
		encryptor:      encryptor,
		shouldCompress: true,
		compressorPool: compressorPool,
		bufferPool:     bufpool.New(),
	}
}

type senderJSON struct {
	client         *resty.Client
	retrierFactory retrymgr.RetrierFactory
	hmacSigner     hmacsigner.HMACSigner
	encryptor      asymcrypt.Encryptor
	shouldCompress bool
	compressorPool *gzippool.WriterPool
	bufferPool     *bufpool.Pool
}

func (s *senderJSON) encryptBody(src *bufpool.Buffer) error {
	if s.encryptor == nil {
		return nil
	}

	ciphertext, err := s.encryptor.Encrypt(src.Bytes())
	if err != nil {
		return err
	}

	// reuse original buffer
	src.Reset()

	n, err := src.Write(ciphertext)
	if err != nil {
		return fmt.Errorf("cannot reuse body buffer: %w", err)
	}

	// sanity check
	if n != len(ciphertext) {
		return fmt.Errorf("incomplete encrypted body written to buffer: written %d, expected %d", n, len(ciphertext))
	}

	return nil
}

func (s *senderJSON) signBody(src *bufpool.Buffer, headers map[string]string) error {
	signature, err := s.hmacSigner.Sign(src.Bytes())
	switch {
	case errors.Is(err, hmacsigner.ErrMissingSecretKey):
		// no secret key, no need to add any headers
	case err == nil:
		headers[httpheaders.HeaderKeyHashSHA256] = httpheaders.GetHashSHA256FromBytes(signature).String()
	default:
		return err
	}

	return nil
}

func (s *senderJSON) compressBody(src *bufpool.Buffer, headers map[string]string) (*bufpool.Buffer, error) {
	buf := s.bufferPool.Get()
	wgz := s.compressorPool.Get(buf)

	_, err := io.Copy(wgz, src)
	if err != nil {
		return src, err
	}

	err = wgz.Close()
	if err != nil {
		return src, err
	}

	headers[httpheaders.HeaderKeyContentEncoding] = httpheaders.ContentEncodingGzip.String()

	return buf, nil
}

func (s *senderJSON) prepareBody(src *bufpool.Buffer) (*bufpool.Buffer, map[string]string, error) {
	headers := make(map[string]string)

	// encrypt before signing to get message authentication regardless of the encryption mechanism.
	err := s.encryptBody(src)
	if err != nil {
		return src, headers, fmt.Errorf("cannot encrypt body: %w", err)
	}

	err = s.signBody(src, headers)
	if err != nil {
		return src, headers, fmt.Errorf("cannot sign body: %w", err)
	}

	if !s.shouldCompress {
		return src, headers, nil
	}

	buf, err := s.compressBody(src, headers)
	if err != nil {
		return src, headers, fmt.Errorf("cannot compress body: %w", err)
	}

	src.Close() // we don't need original body anymore

	return buf, headers, nil
}

func (s *senderJSON) sendSingleRequest(ctx context.Context, method, url string, headers map[string]string, body *bufpool.Buffer) (*resty.Response, error) {
	req := s.client.R().
		SetContext(ctx).
		SetHeader(httpheaders.HeaderKeyContentType, httpheaders.ContentTypeApplicationJSON.String())

	req.Method = method
	req.URL = url

	req.SetHeaders(headers)
	req.SetBody(body.Reader())

	resp, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSenderRequestFailed, err)
	}

	if !resp.IsSuccess() {
		return resp, fmt.Errorf("expected success, got status %v: %w", resp.Status(), ErrSenderResponseNotOK)
	}

	return resp, nil
}

func (s *senderJSON) sendWithRetries(ctx context.Context, method, url string, origBody *bufpool.Buffer) (*resty.Response, error) {
	body, headers, err := s.prepareBody(origBody)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare request body: %w", err)
	}

	defer body.Close()

	return retrymgr.NewRetrier[*resty.Response](s.retrierFactory).Do(
		ctx, "send_metrics_json",
		func(ctx context.Context) (*resty.Response, error) {
			return s.sendSingleRequest(ctx, method, url, headers, body)
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

	buf := s.bufferPool.Get() // the buffer will be closed in [senderJSON.sendWithRetries] method

	if err := json.NewEncoder(buf).Encode(metric); err != nil {
		return fmt.Errorf("json encoder failed: %w", err)
	}

	_, err := s.sendWithRetries(ctx, http.MethodPost, "/update/", buf)

	return err
}

func (s *senderJSON) SendBatch(ctx context.Context, metrics model.MetricSet) (model.MetricSet, error) {
	if metrics.Empty() {
		return model.NewMetricSet(), nil
	}

	buf := s.bufferPool.Get() // the buffer will be closed in [senderJSON.sendWithRetries] method

	if err := json.NewEncoder(buf).Encode(metrics.Values()); err != nil {
		return model.NewMetricSet(), fmt.Errorf("json encoder failed: %w", err)
	}

	resp, err := s.sendWithRetries(ctx, http.MethodPost, "/updates/", buf)
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
