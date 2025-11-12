package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/handler/httpheaders"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner"
	"github.com/bq2cd/yp-go-metrics/pkg/hmacsigner/hmacsignertest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHMACSigner(t *testing.T) {
	type args struct {
		l      log.Logger
		signer hmacsigner.HMACSigner
	}
	type testcase struct {
		args      args
		assertion func(testing.TB, args, Middleware)
	}
	tests := map[string]testcase{
		"default": {
			args: args{
				l:      log.NewNoopLogger(),
				signer: hmacsigner.NewHMACSigner(nil),
			},
			assertion: func(t testing.TB, args args, got Middleware) {
				next := &middlewareHandler{}
				m := got(next)
				require.IsType(t, &middlewareHandler{}, m)
				mh := m.(*middlewareHandler)
				require.IsType(t, &hmacSignerMiddleware{}, mh.impl)
				impl := mh.impl.(*hmacSignerMiddleware)
				assert.Equal(t, args.l, impl.logger)
				assert.Equal(t, args.signer, impl.signer)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := HMACSigner(tt.args.l, tt.args.signer)
			tt.assertion(t, tt.args, got)
		})
	}
}

func Test_hmacSignerMiddleware_Intercept(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		signer func() *hmacsignertest.MockHMACSigner
	}
	type args struct {
		r    func() *http.Request
		next http.Handler
	}
	type want struct {
		status int
		hash   httpheaders.HashSHA256
		body   []byte
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]map[string]testcase{
		"incoming requests, no secret key configured": {
			"any request without hash header is accepted": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(false)
						m.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
			"any request with hash header is accepted, hash header is ignored": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(false)
						m.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.HashSHA256("possibly incorrect hash").Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
		},
		"incoming requests, secret key configured": {
			"requests without body accepted without verification, hash header is ignored": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
						httpheaders.HashSHA256("example hash").Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
			// FIXME:
			// Such requests should not be accepted, but we are forced to accept them
			// because of `go-autotests` which do not sign their requests; see
			// https://github.com/Yandex-Practicum/go-autotests/blob/0591b1dbbcbcf741c41c8eca0718bf676ed7307f/cmd/metricstest_v2/iteration14_test.go#L462
			"requests without hash header are accepted, but response is signed": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
						m.EXPECT().Sign([]byte(`signed response`)).Return([]byte(`response signature`), nil).Times(1)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`signed response`))
					}),
				},
				want: want{
					status: http.StatusOK,
					hash:   httpheaders.GetHashSHA256FromBytes([]byte(`response signature`)),
					body:   []byte(`signed response`),
				},
			},
			"requests with hash header failing verification are not accepted": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify([]byte(`some client data`), []byte(`some invalid signature`)).Return(hmacsigner.ErrSignatureMismatch).Times(1)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.GetHashSHA256FromBytes([]byte(`some invalid signature`)).Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusBadRequest,
					body:   []byte(`signature mismatch` + "\n"),
				},
			},
			"requests with invalid hash header return 500": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify(gomock.Any(), gomock.All()).Times(0)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						r.Header.Set(httpheaders.HeaderKeyHashSHA256, `zzz++=not-a-hex-string-at-all++`)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusInternalServerError,
					body:   []byte{},
				},
			},
			"requests with valid hash header are accepted": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify([]byte(`some client data`), []byte(`some valid signature`)).Return(nil).Times(1)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.GetHashSHA256FromBytes([]byte(`some valid signature`)).Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
		},
		"outgoing responses, no secret key configured": {
			"server responds without hash header": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(false)
						m.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
		},
		"outgoing responses, secret key configured": {
			"server responds without hash header when body is empty and status < 300": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify([]byte(`some client data`), []byte(`some valid signature`)).Return(nil).Times(1)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.GetHashSHA256FromBytes([]byte(`some valid signature`)).Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				},
				want: want{
					status: http.StatusOK,
					body:   []byte{},
				},
			},
			"server responds without hash header when status >= 300": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify([]byte(`some client data`), []byte(`some valid signature`)).Return(nil).Times(1)
						m.EXPECT().Sign(gomock.Any()).Times(0)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.GetHashSHA256FromBytes([]byte(`some valid signature`)).Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Error(w, "some error here", http.StatusNotFound)
					}),
				},
				want: want{
					status: http.StatusNotFound,
					body:   []byte(`some error here` + "\n"),
				},
			},
			"server responds with hash header when body is non-empty and status < 300": {
				fields: fields{
					signer: func() *hmacsignertest.MockHMACSigner {
						m := hmacsignertest.NewMockHMACSigner(ctrl)
						m.EXPECT().HasKey().Return(true)
						m.EXPECT().Verify([]byte(`some client data`), []byte(`some valid signature`)).Return(nil).Times(1)
						m.EXPECT().Sign([]byte(`some response data`)).Return([]byte(`some response signature`), nil).Times(1)
						return m
					},
				},
				args: args{
					r: func() *http.Request {
						r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("some client data"))
						httpheaders.GetHashSHA256FromBytes([]byte(`some valid signature`)).Apply(r.Header)
						return r
					},
					next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`some response data`))
					}),
				},
				want: want{
					status: http.StatusOK,
					hash:   httpheaders.GetHashSHA256FromBytes([]byte(`some response signature`)),
					body:   []byte(`some response data`),
				},
			},
		},
	}
	for gname, cases := range tests {
		t.Run(gname, func(t *testing.T) {
			for name, tt := range cases {
				t.Run(name, func(t *testing.T) {
					// Arrange
					logger := log.NewTestLogger()
					m := &hmacSignerMiddleware{
						logger: logger,
						signer: tt.fields.signer(),
					}
					rw := httptest.NewRecorder()

					// Act
					m.Intercept(rw, tt.args.r(), tt.args.next)

					// Assert
					resp := rw.Result()
					body, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					defer func() { assert.NoError(t, resp.Body.Close()) }()

					assert.Equal(t, tt.want.status, resp.StatusCode)
					assert.Truef(t, tt.want.hash.Matches(resp.Header), "hash header mismatch")
					assert.Equal(t, tt.want.body, body)

					events := logger.RecordedEvents()
					if len(events) > 0 {
						t.Logf("recorded log events: %v", events)
					}
				})
			}
		})
	}
}

func Test_hmacSignerResponseWriter_Write(t *testing.T) {
	type fields struct {
		data *bytes.Buffer
	}
	type args struct {
		data []byte
	}
	type want struct {
		got      int
		wantErr  func(testing.TB, error)
		wantData []byte
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"no data written": {
			fields: fields{data: bytes.NewBuffer([]byte(`123`))},
			args:   args{data: nil},
			want: want{
				got:      0,
				wantErr:  func(t testing.TB, err error) { require.NoError(t, err) },
				wantData: []byte(`123`),
			},
		},
		"some data written": {
			fields: fields{data: bytes.NewBuffer(nil)},
			args:   args{data: []byte(`456`)},
			want: want{
				got:      3,
				wantErr:  func(t testing.TB, err error) { require.NoError(t, err) },
				wantData: []byte(`456`),
			},
		},
		"some data appended": {
			fields: fields{data: bytes.NewBuffer([]byte(`123`))},
			args:   args{data: []byte(`456`)},
			want: want{
				got:      3,
				wantErr:  func(t testing.TB, err error) { require.NoError(t, err) },
				wantData: []byte(`123456`),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hw := &hmacSignerResponseWriter{
				data: tt.fields.data,
			}
			got, err := hw.Write(tt.args.data)
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.got, got)
			assert.Equal(t, tt.want.wantData, hw.data.Bytes())
		})
	}
}

func Test_hmacSignerResponseWriter_WriteHeader(t *testing.T) {
	type fields struct {
		statusCode int
	}
	type args struct {
		statusCode int
	}
	type want struct {
		statusCode int
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"new status code written": {
			fields: fields{},
			args:   args{statusCode: http.StatusAccepted},
			want:   want{statusCode: http.StatusAccepted},
		},
		"old status code overwritten": {
			fields: fields{statusCode: http.StatusAlreadyReported},
			args:   args{statusCode: http.StatusAccepted},
			want:   want{statusCode: http.StatusAccepted},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hw := &hmacSignerResponseWriter{
				statusCode: tt.fields.statusCode,
			}
			hw.WriteHeader(tt.args.statusCode)
			assert.Equal(t, tt.want.statusCode, hw.statusCode)
		})
	}
}
