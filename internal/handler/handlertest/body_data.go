package handlertest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
)

// BodyData contains raw bytes and their content type (as defined in HTTP specification).
type BodyData struct {
	T           testing.TB
	data        []byte
	contentType httpheaders.ContentType
}

// NewBodyData creates an instance of [NewBodyData] without content type.
func NewBodyData(t testing.TB, data []byte) BodyData {
	return BodyData{
		T:           t,
		data:        data,
		contentType: httpheaders.ContentTypeEmpty,
	}
}

// NewBodyDataOfType creates an instance of [BodyData] with provided content type.
func NewBodyDataOfType(t testing.TB, data []byte, contentType httpheaders.ContentType) BodyData {
	return BodyData{
		T:           t,
		data:        data,
		contentType: contentType,
	}
}

// NewBodyDataFromMetric creates an instance of [BodyData] by JSON-encoding of a given [Metric]
// and setting the corresponding content type.
func NewBodyDataFromMetric(t testing.TB, m model.Metric) BodyData {
	t.Helper()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	require.NoError(t, err)
	return BodyData{
		T:           t,
		data:        buf.Bytes(),
		contentType: httpheaders.ContentTypeApplicationJSON,
	}
}

// NewBodyDataFromMetrics creates an instance of [BodyData] by JSON-encoding a slice of [Metric] objects
// and setting the corresponding content type
func NewBodyDataFromMetrics(t testing.TB, metrics []model.Metric) BodyData {
	t.Helper()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(metrics)
	require.NoError(t, err)
	return BodyData{
		T:           t,
		data:        buf.Bytes(),
		contentType: httpheaders.ContentTypeApplicationJSON,
	}
}

// NewBodyDataFromMetricKey creates an instance of [BodyData] by JSON-encoding of a given [MetricKey]
// and setting the corresponding content type.
func NewBodyDataFromMetricKey(t testing.TB, k model.MetricKey) BodyData {
	t.Helper()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(k)
	require.NoError(t, err)
	return BodyData{
		T:           t,
		data:        buf.Bytes(),
		contentType: httpheaders.ContentTypeApplicationJSON,
	}
}

// NewBodyDataFromResponse creates an instance of [BodyData] from HTTP response.
// Content type is extracted from the response headers and compressed data is
// decompressed.
func NewBodyDataFromResponse(t testing.TB, resp *http.Response) BodyData {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

	if httpheaders.ContentEncodingGzip.Matches(resp.Header) {
		r := bytes.NewReader(body)
		rgz, err := gzip.NewReader(r)
		require.NoError(t, err)
		body, err = io.ReadAll(rgz)
		require.NoError(t, err)
		err = rgz.Close()
		require.NoError(t, err)
	}

	return BodyData{
		T:           t,
		data:        body,
		contentType: httpheaders.GetContentType(resp.Header),
	}
}

// AsType creates a shallow copy of [BodyData] with different content type.
// The data bytes are shared with the original instance.
func (b BodyData) AsType(contentType httpheaders.ContentType) BodyData {
	return BodyData{
		T:           b.T,
		data:        b.data,
		contentType: contentType,
	}
}

// TransformData creates a new instance of [BodyData] by copying original data and applying transformation function
// to the copy. Original instance of [BodyData] remains untouched.
func (b BodyData) TransformData(transformFn func([]byte) []byte) BodyData {
	transformed := b.data
	if transformFn != nil {
		transformed = transformFn(b.data)
	}

	return BodyData{
		T:           b.T,
		data:        transformed,
		contentType: b.contentType,
	}
}

// NewReader creates a [io.ReadCloser] object suitable for usage in HTTP request.
// The data is compressed if `shouldCompress` is `true`.
func (b *BodyData) NewReader(shouldCompress bool) io.ReadCloser {
	b.T.Helper()
	var body io.ReadCloser = http.NoBody
	if len(b.data) == 0 {
		return body
	}
	r := bytes.NewReader(b.data)
	if !shouldCompress {
		return io.NopCloser(r)
	}
	var buf bytes.Buffer
	wgz := gzip.NewWriter(&buf)
	_, err := io.Copy(wgz, r)
	require.NoError(b.T, err)
	err = wgz.Close()
	require.NoError(b.T, err)
	return io.NopCloser(&buf)
}

// NewRequest creates new [http.Request] for given method and url, using the underlying
// data as the request's body. The data is compressed if `shouldCompress` is `true`.
func (b *BodyData) NewRequest(method, url string, shouldCompress bool) *http.Request {
	b.T.Helper()
	body := b.NewReader(shouldCompress)
	req, err := http.NewRequest(method, url, body)
	require.NoError(b.T, err)
	if shouldCompress {
		httpheaders.ContentEncodingGzip.Apply(req.Header)
	}
	b.contentType.Apply(req.Header)
	return req
}

// GetDataSignature calculates HMAC signature of the underlying data bytes using provided [hmacsigner.HMACSigner].
func (b *BodyData) GetDataSignature(signer hmacsigner.HMACSigner) httpheaders.HashSHA256 {
	b.T.Helper()
	require.NotNil(b.T, signer)
	if !signer.HasKey() {
		return httpheaders.HashSHA256Empty
	}
	signature, err := signer.Sign(b.data)
	require.NoError(b.T, err)
	return httpheaders.GetHashSHA256FromBytes(signature)
}

// Len returns the length of the underlying data bytes.
func (b *BodyData) Len() int {
	return len(b.data)
}

// AssertData compares underlying data bytes with the provided bytes using [assert.Assertions]
// from `testify` library. The content type is taken into account: for some content types (e.g. JSON)
// more specific assertion is used, otherwise [assert.Equal].
func (b *BodyData) AssertData(expected []byte) {
	b.T.Helper()
	switch b.contentType {
	case httpheaders.ContentTypeApplicationJSON:
		assert.JSONEq(b.T, string(expected), string(b.data))
	default:
		assert.Equal(b.T, expected, b.data)
	}
}

// AssertType compares underlying content type with the provided content type using [assert.Assertions]
// from `testify` library.
func (b *BodyData) AssertType(expected httpheaders.ContentType) {
	b.T.Helper()
	assert.Equal(b.T, expected, b.contentType)
}

// AssertEqual compares current [BodyData] object with the other [BodyData] object using [assert.Assertions]
// from `testify` library.
// The objects are considered equal if both their content types and data bytes are equal.
func (b *BodyData) AssertEqual(other BodyData) {
	b.T.Helper()
	b.AssertType(other.contentType)
	b.AssertData(other.data)
}

// DecodeBodyDataAsJSON attempts to JSON-decode data bytes of provided [BodyData] object.
// It does not check content type at all - if bytes are valid JSON, it will decode them,
// otherwise it will return an error from JSON decoder.
func DecodeBodyDataAsJSON[T any](body *BodyData) (T, error) {
	var target T

	err := json.Unmarshal(body.data, &target)

	return target, err
}
