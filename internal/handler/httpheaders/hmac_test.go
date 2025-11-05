package httpheaders

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHashSHA256(t *testing.T) {
	type args struct {
		header http.Header
	}
	type want struct {
		got HashSHA256
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty header return empty hash": {
			args: args{
				header: func() http.Header {
					h := http.Header{}
					return h
				}(),
			},
			want: want{got: HashSHA256Empty},
		},
		"some header return some hash": {
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Set(HeaderKeyHashSHA256, "123")
					return h
				}(),
			},
			want: want{got: HashSHA256("123")},
		},
		"only first header value is returned": {
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add(HeaderKeyHashSHA256, "123")
					h.Add(HeaderKeyHashSHA256, "456")
					return h
				}(),
			},
			want: want{got: HashSHA256("123")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetHashSHA256(tt.args.header)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestGetHashSHA256FromBytes(t *testing.T) {
	type args struct {
		data []byte
	}
	type want struct {
		got HashSHA256
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty bytes result in empty hash": {
			args: args{data: nil},
			want: want{got: HashSHA256Empty},
		},
		"simple message creates hex-encoded hash": {
			args: args{data: []byte(`pellentesque`)},
			want: want{got: HashSHA256("70656c6c656e746573717565")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetHashSHA256FromBytes(tt.args.data)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestHashSHA256_String(t *testing.T) {
	type want struct {
		got string
	}
	type testcase struct {
		h    HashSHA256
		want want
	}
	tests := map[string]testcase{
		"empty hash return empty string": {
			h:    HashSHA256Empty,
			want: want{got: ``},
		},
		"simply returns underlying string": {
			h:    HashSHA256(`msg-123`),
			want: want{got: `msg-123`},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.h.String()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestHashSHA256_Bytes(t *testing.T) {
	type want struct {
		got     []byte
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		h    HashSHA256
		want want
	}
	tests := map[string]testcase{
		"empty hash return empty bytes": {
			h: HashSHA256Empty,
			want: want{
				got: []byte{},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"valid hex string returns decoded bytes": {
			h: HashSHA256("437562696c69612063757261652068616320686162697461737365"),
			want: want{
				got: []byte(`Cubilia curae hac habitasse`),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"invalid hex string returns error": {
			h: HashSHA256("zzz"),
			want: want{
				got: []byte{},
				wantErr: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "invalid byte")
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.h.Bytes()
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestHashSHA256_Apply(t *testing.T) {
	type args struct {
		header http.Header
	}
	type want struct {
		got    HashSHA256
		header http.Header
	}
	type testcase struct {
		h    HashSHA256
		args args
		want want
	}
	tests := map[string]testcase{
		"empty hash does not change empty header": {
			h: HashSHA256Empty,
			args: args{
				header: func() http.Header {
					return http.Header{}
				}(),
			},
			want: want{
				got:    HashSHA256Empty,
				header: http.Header{},
			},
		},
		"empty hash removes existing header": {
			h: HashSHA256Empty,
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add(HeaderKeyHashSHA256, "123")
					h.Add(HeaderKeyHashSHA256, "456")
					return h
				}(),
			},
			want: want{
				got:    HashSHA256Empty,
				header: http.Header{},
			},
		},
		"some hash is added to empty header": {
			h: HashSHA256("123"),
			args: args{
				header: func() http.Header {
					return http.Header{}
				}(),
			},
			want: want{
				got:    HashSHA256("123"),
				header: http.Header{HeaderKeyHashSHA256: []string{"123"}},
			},
		},
		"some hash overwrites existing header": {
			h: HashSHA256("789"),
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add(HeaderKeyHashSHA256, "123")
					h.Add(HeaderKeyHashSHA256, "456")
					return h
				}(),
			},
			want: want{
				got:    HashSHA256("789"),
				header: http.Header{HeaderKeyHashSHA256: []string{"789"}},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.h.Apply(tt.args.header)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestHashSHA256_Matches(t *testing.T) {
	type args struct {
		header http.Header
	}
	type want struct {
		got bool
	}
	type testcase struct {
		h    HashSHA256
		args args
		want want
	}
	tests := map[string]testcase{
		"empty header matches empty hash": {
			h:    HashSHA256Empty,
			args: args{header: http.Header{}},
			want: want{got: true},
		},
		"non-empty header with hash matches empty hash": {
			h: HashSHA256Empty,
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add("Key 1", "Value 1")
					h.Add("Key 1", "Value 2")
					return h
				}(),
			},
			want: want{got: true},
		},
		"non-empty header with hash does not match different hash": {
			h: HashSHA256("123"),
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add(HeaderKeyHashSHA256, "456")
					return h
				}(),
			},
			want: want{got: false},
		},
		"non-empty header with hash matches same hash": {
			h: HashSHA256("123"),
			args: args{
				header: func() http.Header {
					h := http.Header{}
					h.Add(HeaderKeyHashSHA256, "123")
					h.Add(HeaderKeyHashSHA256, "456")
					return h
				}(),
			},
			want: want{got: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.h.Matches(tt.args.header)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
