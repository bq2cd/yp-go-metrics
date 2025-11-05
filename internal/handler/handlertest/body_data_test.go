package handlertest

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewBodyData(t *testing.T) {
	type args struct {
		t    *testing.T
		data []byte
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty data": {
			args: args{t: t, data: nil},
			want: want{got: BodyData{T: t}},
		},
		"some data": {
			args: args{t: t, data: []byte(`123`)},
			want: want{got: BodyData{T: t, data: []byte(`123`)}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewBodyData(tt.args.t, tt.args.data)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNewBodyDataOfType(t *testing.T) {
	type args struct {
		t           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"no content type": {
			args: args{t: t, data: []byte(`123`)},
			want: want{got: BodyData{T: t, data: []byte(`123`)}},
		},
		"some content type": {
			args: args{t: t, data: []byte(`123`), contentType: httpheaders.ContentTypeApplicationJSON},
			want: want{got: BodyData{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewBodyDataOfType(tt.args.t, tt.args.data, tt.args.contentType)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestNewBodyDataFromMetric(t *testing.T) {
	type args struct {
		t *testing.T
		m model.Metric
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"some counter": {
			args: args{t: t, m: model.NewCounterMetric("id1", -123)},
			want: want{got: BodyData{T: t, data: []byte(`{"id": "id1", "type": "counter", "delta": -123}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
		"some gauge": {
			args: args{t: t, m: model.NewGaugeMetric("id1", -1.23)},
			want: want{got: BodyData{T: t, data: []byte(`{"id": "id1", "type": "gauge", "value": -1.23}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewBodyDataFromMetric(tt.args.t, tt.args.m)
			assert.Equal(t, tt.want.got.contentType, got.contentType)
			assert.JSONEq(t, string(tt.want.got.data), string(got.data))
		})
	}
}

func TestNewBodyDataFromMetricKey(t *testing.T) {
	type args struct {
		t *testing.T
		k model.MetricKey
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"some counter": {
			args: args{t: t, k: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: want{got: BodyData{T: t, data: []byte(`{"id": "id1", "type": "counter"}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
		"some gauge": {
			args: args{t: t, k: model.NewMetricKey(model.MetricTypeGauge, "id1")},
			want: want{got: BodyData{T: t, data: []byte(`{"id": "id1", "type": "gauge"}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewBodyDataFromMetricKey(tt.args.t, tt.args.k)
			assert.Equal(t, tt.want.got.contentType, got.contentType)
			assert.JSONEq(t, string(tt.want.got.data), string(got.data))
		})
	}
}

func TestNewBodyDataFromResponse(t *testing.T) {
	type args struct {
		t    *testing.T
		resp func() *http.Response
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty response": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(``), contentType: httpheaders.ContentTypeEmpty}},
		},
		"data without content type": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					w.WriteHeader(http.StatusContinue)
					_, err := w.Write([]byte(`123`))
					require.NoError(t, err)
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeEmpty}},
		},
		"no explicit header write results in plain content type": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					_, err := w.Write([]byte(`123`))
					require.NoError(t, err)
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextPlain.UTF8()}},
		},
		"some json": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
					w.WriteHeader(http.StatusContinue)
					_, err := w.Write([]byte(`{}`))
					require.NoError(t, err)
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(`{}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
		"json with invalid content type": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					_, err := w.Write([]byte(`{}`))
					require.NoError(t, err)
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(`{}`), contentType: httpheaders.ContentTypeTextPlain.UTF8()}},
		},
		"compressed json": {
			args: args{
				t: t,
				resp: func() *http.Response {
					w := httptest.NewRecorder()
					httpheaders.ContentTypeApplicationJSON.Apply(w.Header())
					httpheaders.ContentEncodingGzip.Apply(w.Header())
					w.WriteHeader(http.StatusContinue)
					wgz := gzip.NewWriter(w)
					_, err := wgz.Write([]byte(`{"k1": "v1"}`))
					require.NoError(t, err)
					err = wgz.Close()
					require.NoError(t, err)
					return w.Result()
				},
			},
			want: want{got: BodyData{T: t, data: []byte(`{"k1": "v1"}`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := tt.args.resp()
			defer func() { _ = resp.Body.Close() }()
			got := NewBodyDataFromResponse(tt.args.t, resp)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestBodyData_AsType(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		contentType httpheaders.ContentType
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"adding content type": {
			fields: fields{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeEmpty},
			args:   args{contentType: httpheaders.ContentTypeTextHTML},
			want:   want{got: BodyData{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextHTML}},
		},
		"overriding content type": {
			fields: fields{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextHTML},
			args:   args{contentType: httpheaders.ContentTypeApplicationJSON},
			want:   want{got: BodyData{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeApplicationJSON}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			got := b.AsType(tt.args.contentType)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestBodyData_NewReader(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		shouldCompress bool
	}
	type want struct {
		data []byte
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"no data": {
			fields: fields{T: t, data: []byte(``), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{shouldCompress: false},
			want:   want{data: []byte(``)},
		},
		"no compression": {
			fields: fields{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{shouldCompress: false},
			want:   want{data: []byte(`123`)},
		},
		"with compression": {
			fields: fields{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{shouldCompress: true},
			want:   want{data: []byte(`123`)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			got := b.NewReader(tt.args.shouldCompress)
			var (
				data    []byte
				errRead error
			)
			if !tt.args.shouldCompress {
				data, errRead = io.ReadAll(got)
			} else {
				rgz, err := gzip.NewReader(got)
				require.NoError(t, err)
				data, errRead = io.ReadAll(rgz)
			}
			require.NoError(t, errRead)
			assert.Equal(t, tt.want.data, data)
		})
	}
}

func TestBodyData_NewRequest(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		method         string
		url            string
		shouldCompress bool
	}
	type want struct {
		method      string
		url         string
		data        []byte
		contentType httpheaders.ContentType
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"no compression": {
			fields: fields{T: t, data: []byte(`123`), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{method: http.MethodGet, url: "/", shouldCompress: false},
			want:   want{method: http.MethodGet, url: "/", data: []byte(`123`), contentType: httpheaders.ContentTypeTextPlain},
		},
		"with compression": {
			fields: fields{T: t, data: []byte(`{}`), contentType: httpheaders.ContentTypeApplicationJSON},
			args:   args{method: http.MethodPost, url: "/update", shouldCompress: true},
			want:   want{method: http.MethodPost, url: "/update", data: []byte(`{}`), contentType: httpheaders.ContentTypeApplicationJSON},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			got := b.NewRequest(tt.args.method, tt.args.url, tt.args.shouldCompress)

			assert.Equal(t, tt.want.method, got.Method)
			assert.Equal(t, tt.want.url, got.URL.Path)
			assert.Truef(t, tt.want.contentType.Matches(got.Header), "expected content type %v, got %v", tt.fields.contentType, httpheaders.GetContentType(got.Header))

			var (
				data    []byte
				errRead error
			)
			if !tt.args.shouldCompress {
				data, errRead = io.ReadAll(got.Body)
			} else {
				rgz, err := gzip.NewReader(got.Body)
				require.NoError(t, err)
				data, errRead = io.ReadAll(rgz)
			}
			require.NoError(t, errRead)
			assert.Equal(t, tt.want.data, data)
		})
	}
}

func TestBodyData_Len(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type want struct {
		got int
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"no data": {
			fields: fields{T: t, data: []byte(``), contentType: httpheaders.ContentTypeTextPlain},
			want:   want{got: 0},
		},
		"some data": {
			fields: fields{T: t, data: []byte(`12345`), contentType: httpheaders.ContentTypeTextPlain},
			want:   want{got: 5},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			got := b.Len()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestBodyData_AssertData(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		expected []byte
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		"no data": {
			fields: fields{T: t, data: []byte(``), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{expected: []byte(``)},
		},
		"some data": {
			fields: fields{T: t, data: []byte(`12345`), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{expected: []byte(`12345`)},
		},
		"some json": {
			fields: fields{T: t, data: []byte(`{"k1": "v1", "k2": "v2", "k3": -1}`), contentType: httpheaders.ContentTypeApplicationJSON},
			args:   args{expected: []byte(`{"k3":-1, "k1":"v1", "k2":"v2"}`)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			b.AssertData(tt.args.expected)
		})
	}
}

func TestBodyData_AssertType(t *testing.T) {
	type fields struct {
		T           *testing.T
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		expected httpheaders.ContentType
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		"no type": {
			fields: fields{T: t, data: []byte(``)},
			args:   args{expected: httpheaders.ContentTypeEmpty},
		},
		"some type": {
			fields: fields{T: t, data: []byte(``), contentType: httpheaders.ContentTypeTextPlain},
			args:   args{expected: httpheaders.ContentTypeTextPlain},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			b.AssertType(tt.args.expected)
		})
	}
}

func TestBodyData_AssertEqual(t *testing.T) {
	type fields struct {
		T           testing.TB
		data        []byte
		contentType httpheaders.ContentType
	}
	type args struct {
		other BodyData
	}
	type want struct {
	}
	type testcase struct {
		fields fields
		args   args
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:           tt.fields.T,
				data:        tt.fields.data,
				contentType: tt.fields.contentType,
			}
			b.AssertEqual(tt.args.other)
		})
	}
}

func TestNewBodyDataFromMetrics(t *testing.T) {
	type args struct {
		t       testing.TB
		metrics []model.Metric
	}
	type want struct {
		got BodyData
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewBodyDataFromMetrics(tt.args.t, tt.args.metrics)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestBodyData_GetDataSignature(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		data []byte
	}
	type args struct {
		signer hmacsigner.HMACSigner
	}
	type want struct {
		got httpheaders.HashSHA256
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"no signature generated without secret key": {
			fields: fields{data: []byte(`123`)},
			args: args{
				signer: func() hmacsigner.HMACSigner {
					m := hmacsignertest.NewMockHMACSigner(ctrl)
					m.EXPECT().HasKey().Return(false)
					m.EXPECT().Sign(gomock.Any()).Times(0)
					return m
				}(),
			},
			want: want{got: httpheaders.HashSHA256Empty},
		},
		"some signature generated with secret key": {
			fields: fields{data: []byte(`123`)},
			args: args{
				signer: func() hmacsigner.HMACSigner {
					m := hmacsignertest.NewMockHMACSigner(ctrl)
					m.EXPECT().HasKey().Return(true)
					m.EXPECT().Sign([]byte(`123`)).Return([]byte(`some signature`), nil)
					return m
				}(),
			},
			want: want{got: httpheaders.HashSHA256("736f6d65207369676e6174757265")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &BodyData{
				T:    t,
				data: tt.fields.data,
			}
			got := b.GetDataSignature(tt.args.signer)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
