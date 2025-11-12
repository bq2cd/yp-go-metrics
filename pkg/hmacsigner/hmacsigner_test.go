package hmacsigner

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHMACSigner(t *testing.T) {
	type args struct {
		secretKey []byte
	}
	type want struct {
		got *hmacSigner
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"given key is stored internally": {
			args: args{secretKey: []byte(`123`)},
			want: want{got: &hmacSigner{secretKey: []byte(`123`)}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewHMACSigner(tt.args.secretKey)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_hmacSigner_Sign(t *testing.T) {
	type fields struct {
		secretKey []byte
	}
	type args struct {
		signingMessage []byte
	}
	type want struct {
		got     []byte
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty key results in error": {
			fields: fields{secretKey: []byte{}},
			args:   args{signingMessage: []byte(`msg`)},
			want: want{
				got: nil,
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrMissingSecretKey)
				},
			},
		},
		"empty message is returned as is": {
			fields: fields{secretKey: []byte(`123`)},
			args:   args{signingMessage: nil},
			want: want{
				got: []byte{},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"simple message signed": {
			fields: fields{secretKey: []byte(`123`)},
			args:   args{signingMessage: []byte(`msg`)},
			want: want{
				got: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`27c3dad37f03c56c7bfa5997cd3cbaec47e46a6641b344306d891e61f6ea339d`)
					require.NoError(t, err)
					return data
				}(),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"another simple message signed": {
			fields: fields{secretKey: []byte(`consectetur`)},
			args:   args{signingMessage: []byte(`faucibus`)},
			want: want{
				got: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`1ae431c428a7aa90f49beef9854d9b8b9443196d806462e95627bf81761edda5`)
					require.NoError(t, err)
					return data
				}(),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hs := &hmacSigner{
				secretKey: tt.fields.secretKey,
			}
			got, err := hs.Sign(tt.args.signingMessage)
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_hmacSigner_Verify(t *testing.T) {
	type fields struct {
		secretKey []byte
	}
	type args struct {
		signedMessage []byte
		signature     []byte
	}
	type want struct {
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty key results in error": {
			fields: fields{secretKey: []byte{}},
			args: args{
				signedMessage: []byte(`msg`),
				signature: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`27c3dad37f03c56c7bfa5997cd3cbaec47e46a6641b344306d891e61f6ea339d`)
					require.NoError(t, err)
					return data
				}(),
			},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrMissingSecretKey)
				},
			},
		},
		"empty message and empty signature return OK": {
			fields: fields{secretKey: []byte(`123`)},
			args:   args{},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"simple message is not verified by wrong signature": {
			fields: fields{secretKey: []byte(`123`)},
			args: args{
				signedMessage: []byte(`msg1`),
				signature: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`27c3dad37f03c56c7bfa5997cd3cbaec47e46a6641b344306d891e61f6ea339d`)
					require.NoError(t, err)
					return data
				}(),
			},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrSignatureMismatch)
				},
			},
		},
		"simple message is verified by correct signature": {
			fields: fields{secretKey: []byte(`123`)},
			args: args{
				signedMessage: []byte(`msg`),
				signature: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`27c3dad37f03c56c7bfa5997cd3cbaec47e46a6641b344306d891e61f6ea339d`)
					require.NoError(t, err)
					return data
				}(),
			},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"another simple message is verified by correct signature": {
			fields: fields{secretKey: []byte(`consectetur`)},
			args: args{
				signedMessage: []byte(`faucibus`),
				signature: func() []byte {
					// generated at https://www.freeformatter.com/hmac-generator.html#before-output
					data, err := hex.DecodeString(`1ae431c428a7aa90f49beef9854d9b8b9443196d806462e95627bf81761edda5`)
					require.NoError(t, err)
					return data
				}(),
			},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hs := &hmacSigner{
				secretKey: tt.fields.secretKey,
			}
			err := hs.Verify(tt.args.signedMessage, tt.args.signature)
			tt.want.wantErr(t, err)
		})
	}
}

func Test_hmacSigner_HasKey(t *testing.T) {
	type fields struct {
		secretKey []byte
	}
	type want struct {
		got bool
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"no key": {
			fields: fields{secretKey: nil},
			want:   want{got: false},
		},
		"some key": {
			fields: fields{secretKey: []byte(`123`)},
			want:   want{got: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hs := &hmacSigner{
				secretKey: tt.fields.secretKey,
			}
			got := hs.HasKey()
			assert.Equal(t, tt.want.got, got)
		})
	}
}
