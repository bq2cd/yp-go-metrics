package repository

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLStorage(t *testing.T) {
	type args struct {
		dbURL url.URL
	}
	type want struct {
		assertion func(*testing.T, *sqlStorage, error)
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty url returns no error": {
			args: args{dbURL: url.URL{}},
			want: want{assertion: func(t *testing.T, got *sqlStorage, err error) {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}},
		},
		"unsupported database type fails": {
			args: args{dbURL: url.URL{Scheme: "sqlite", Path: "/some.db"}},
			want: want{assertion: func(t *testing.T, got *sqlStorage, err error) {
				require.ErrorIs(t, err, ErrUnsupportedDBType)
				assert.Nil(t, got)
			}},
		},
		"valid url succeeds": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:5432", Path: "/db1"}},
			want: want{assertion: func(t *testing.T, got *sqlStorage, err error) {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}},
		}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewSQLStorage(tt.args.dbURL)
			tt.want.assertion(t, got, err)
		})
	}
}

func Test_sqlStorage_Ping(t *testing.T) {
	type fields struct {
		mockDelay time.Duration
		mockErr   error
	}
	type args struct {
		timeout time.Duration
	}
	type want struct {
		wantErr bool
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"database responds fast and without error": {
			fields: fields{mockDelay: 5 * time.Millisecond, mockErr: nil},
			args:   args{timeout: 10 * time.Millisecond},
			want:   want{wantErr: false},
		},
		"database responds slow, context cancelled": {
			fields: fields{mockDelay: 15 * time.Millisecond, mockErr: nil},
			args:   args{timeout: 10 * time.Millisecond},
			want:   want{wantErr: true},
		},
		"database unreachable": {
			fields: fields{mockDelay: 2 * time.Millisecond, mockErr: errors.New("database unreachable")},
			args:   args{timeout: 10 * time.Millisecond},
			want:   want{wantErr: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectPing().WillDelayFor(tt.fields.mockDelay).WillReturnError(tt.fields.mockErr)
			s := &sqlStorage{
				db: db,
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.args.timeout)
			defer cancel()

			// Act
			err = s.Ping(ctx)

			// Assert
			if tt.want.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_sqlStorage_Close(t *testing.T) {
	type fields struct {
		mockErr error
	}
	type want struct {
		wantErr bool
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		"successful close returns no error": {
			fields: fields{mockErr: nil},
			want:   want{wantErr: false},
		},
		"close fails with error": {
			fields: fields{mockErr: errors.New("close error")},
			want:   want{wantErr: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectClose().WillReturnError(tt.fields.mockErr)
			s := &sqlStorage{
				db: db,
			}

			// Act
			err = s.Close()

			// Assert
			if tt.want.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
