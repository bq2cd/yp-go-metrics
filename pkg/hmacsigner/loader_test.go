package hmacsigner

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-load-secret-key-*")
	require.NoError(t, err)
	if len(data) > 0 {
		_, err = f.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, f.Close())
	return f.Name()
}

type tempfileManager []string

func (tfm *tempfileManager) Add(path string) string {
	*tfm = append(*tfm, path)
	return path
}

func (tfm *tempfileManager) RemoveAll() {
	for _, path := range *tfm {
		os.Remove(path)
	}
}

func generateRandomString(t *testing.T, size int) string {
	t.Helper()
	tmp := make([]byte, size)
	rand.Read(tmp)
	return string(tmp)
}

func TestLoadSecretKey(t *testing.T) {
	tempMgr := make(tempfileManager, 0)
	defer tempMgr.RemoveAll()
	bigString := generateRandomString(t, 1_000_000)
	type args struct {
		source string
	}
	type want struct {
		got []byte
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty string -> nil bytes": {
			args: args{source: ""},
			want: want{got: nil},
		},
		"base64-encoded string -> decoded bytes": {
			args: args{source: "NDU2"},
			want: want{got: []byte(`456`)},
		},
		"non-existent file path -> original string as bytes": {
			args: args{source: "/non-existent-123-gfedf"},
			want: want{got: []byte(`/non-existent-123-gfedf`)},
		},
		"some random big string -> original string as bytes": {
			args: args{source: bigString},
			want: want{got: []byte(bigString)},
		},
		"existing empty file -> empty bytes": {
			args: args{
				source: tempMgr.Add(createTempFile(t, nil)),
			},
			want: want{got: []byte{}},
		},
		"existing non-empty file -> file contents": {
			args: args{
				source: tempMgr.Add(createTempFile(t, []byte(`123456`))),
			},
			want: want{got: []byte(`123456`)},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := LoadSecretKey(tt.args.source)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
