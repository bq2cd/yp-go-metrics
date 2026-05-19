package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt"
	"github.com/bq2cd/yp-go-metrics/pkg/asymcrypt/asymcrypttest"
	"github.com/bq2cd/yp-go-metrics/pkg/log"
)

func Test_requestDecryptor_Intercept(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mirrorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.Copy(w, r.Body)
	}

	type want struct {
		status int
		body   []byte
	}
	type testcase struct {
		decryptor func() asymcrypt.Decryptor
		next      http.HandlerFunc
		req       func() *http.Request
		want      want
	}

	tests := map[string]testcase{
		"nil decryptor passes all requests untouched": {
			decryptor: func() asymcrypt.Decryptor { return nil },
			next:      mirrorHandler,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`clear text data`))
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`clear text data`),
			},
		},
		"decryptor decrypts ciphertext successfully": {
			decryptor: func() asymcrypt.Decryptor {
				m := asymcrypttest.NewMockDecryptor(ctrl)
				m.EXPECT().Decrypt([]byte(`ciphertext from request`)).Return([]byte(`decrypted clear text`), nil)
				return m
			},
			next: mirrorHandler,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`ciphertext from request`))
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`decrypted clear text`),
			},
		},
		"decryptor fails to decrypt ciphertext (e.g. incorrect key)": {
			decryptor: func() asymcrypt.Decryptor {
				m := asymcrypttest.NewMockDecryptor(ctrl)
				m.EXPECT().Decrypt([]byte(`ciphertext from request`)).Return(nil, fmt.Errorf("decryption failed"))
				return m
			},
			next: mirrorHandler,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`ciphertext from request`))
			},
			want: want{
				status: http.StatusBadRequest,
				body:   []byte{},
			},
		},
		"middleware cannot read request body": {
			decryptor: func() asymcrypt.Decryptor {
				m := asymcrypttest.NewMockDecryptor(ctrl)
				return m
			},
			next: mirrorHandler,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", &faultyReader{data: []byte(`ciphertext from request`)})
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte{},
			},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			logger := log.NewTestLogger()

			rw := httptest.NewRecorder()
			m := RequestDecryptor(logger, tc.decryptor())(tc.next)

			// Act
			m.ServeHTTP(rw, tc.req())

			// Assert
			resp := rw.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()

			assert.Equal(t, tc.want.status, resp.StatusCode)
			assert.Equal(t, tc.want.body, body)
		})
	}
}

type faultyReader struct {
	data []byte
}

func (r *faultyReader) Read(p []byte) (int, error) {
	n := min(3, len(r.data))
	copy(p[:n], r.data[:n])

	return n, fmt.Errorf("read error")
}
