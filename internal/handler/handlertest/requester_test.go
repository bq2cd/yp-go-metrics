package handlertest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequester(t *testing.T) {
	type args struct {
		t      testing.TB
		client *http.Client
	}
	type want struct {
		got *Requester
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
			got := NewRequester(tt.args.t, tt.args.client)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestRequester_Do(t *testing.T) {
	type fields struct {
		T      testing.TB
		client *http.Client
	}
	type args struct {
		method         string
		url            string
		body           BodyData
		shouldCompress bool
	}
	type want struct {
		got *Response
		err error
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rqr := &Requester{
				T:      tt.fields.T,
				client: tt.fields.client,
			}
			got, err := rqr.Do(tt.args.method, tt.args.url, tt.args.body, tt.args.shouldCompress)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
