package handlertest

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type BodyData struct {
	T           testing.TB
	data        []byte
	contentType httpheaders.ContentType
}

func NewBodyData(t testing.TB, data []byte) BodyData {
	return BodyData{
		T:           t,
		data:        data,
		contentType: httpheaders.ContentTypeEmpty,
	}
}

func NewBodyDataOfType(t testing.TB, data []byte, contentType httpheaders.ContentType) BodyData {
	return BodyData{
		T:           t,
		data:        data,
		contentType: contentType,
	}
}

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

func (b BodyData) AsType(contentType httpheaders.ContentType) BodyData {
	return BodyData{
		T:           b.T,
		data:        b.data,
		contentType: contentType,
	}
}

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

func (b *BodyData) Len() int {
	return len(b.data)
}

func (b *BodyData) AssertData(expected []byte) {
	b.T.Helper()
	switch b.contentType {
	case httpheaders.ContentTypeApplicationJSON:
		assert.JSONEq(b.T, string(expected), string(b.data))
	default:
		assert.Equal(b.T, expected, b.data)
	}
}

func (b *BodyData) AssertType(expected httpheaders.ContentType) {
	b.T.Helper()
	assert.Equal(b.T, expected, b.contentType)
}

func (b *BodyData) AssertEqual(other BodyData) {
	b.T.Helper()
	b.AssertType(other.contentType)
	b.AssertData(other.data)
}
