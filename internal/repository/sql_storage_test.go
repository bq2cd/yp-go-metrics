package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLStorage(t *testing.T) {
	type args struct {
		dbURL url.URL
	}
	type want struct {
		assertConfig  func(*testing.T, *dbconfig.Config, error)
		assertStorage func(*testing.T, *sqlStorage, error)
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"empty url returns no error": {
			args: args{dbURL: url.URL{}},
			want: want{
				assertConfig: func(t *testing.T, got *dbconfig.Config, err error) {
					require.NoError(t, err)
					t.Logf("config: %#v, %v", got, got.Enabled())
					assert.False(t, got.Enabled())
				},
				assertStorage: func(t *testing.T, got *sqlStorage, err error) {
					require.NoError(t, err)
					assert.NotNil(t, got)
				},
			},
		},
		"unsupported database type fails": {
			args: args{dbURL: url.URL{Scheme: "sqlite", Path: "/some.db"}},
			want: want{
				assertConfig: func(t *testing.T, got *dbconfig.Config, err error) {
					require.ErrorIs(t, err, dbconfig.ErrUnsupportedDBType)
					assert.False(t, got.Enabled())
				},
				assertStorage: func(t *testing.T, got *sqlStorage, err error) {
					require.Errorf(t, err, "sql: unknown driver")
					assert.Nil(t, got)
				},
			},
		},
		"valid url succeeds": {
			args: args{dbURL: url.URL{Scheme: "postgres", Host: "localhost:5432", Path: "/db1"}},
			want: want{
				assertConfig: func(t *testing.T, got *dbconfig.Config, err error) {
					require.NoError(t, err)
					assert.True(t, got.Enabled())
				},
				assertStorage: func(t *testing.T, got *sqlStorage, err error) {
					require.NoError(t, err)
					assert.NotNil(t, got)
				},
			},
		}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cfg, err := dbconfig.New(tt.args.dbURL)
			tt.want.assertConfig(t, &cfg, err)
			got, err := NewSQLStorage(cfg)
			tt.want.assertStorage(t, got, err)
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
			require.NoError(t, mock.ExpectationsWereMet())
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
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_sqlStorage_Get(t *testing.T) {
	type fields struct {
		expectMockCall bool
		mockErr        error
		mockDelay      time.Duration
		mockQuery      string
		mockArgs       []driver.Value
		mockRows       func() *sqlmock.Rows
	}
	type args struct {
		key model.MetricKey
	}
	type want struct {
		got     model.Metric
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		timeout time.Duration
		fields  fields
		args    args
		want    want
	}
	tests := map[string]testcase{
		"empty storage returns metric not found error": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `SELECT value FROM metrics_counter WHERE metric_id = $1`,
				mockArgs:       []driver.Value{"id1"},
				mockRows:       func() *sqlmock.Rows { return sqlmock.NewRows(nil) },
			},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: want{
				got: model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrMetricNotFound)
				},
			},
		},
		"counter metric returned": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `SELECT value FROM metrics_counter WHERE metric_id = $1`,
				mockArgs:       []driver.Value{"id1"},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"value"})
					rows.AddRow(123)
					return rows
				},
			},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: want{
				got: model.NewCounterMetric("id1", 123),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"gauge metric returned": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `SELECT value FROM metrics_gauge WHERE metric_id = $1`,
				mockArgs:       []driver.Value{"id1"},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"value"})
					rows.AddRow(-1.23)
					return rows
				},
			},
			args: args{key: model.NewMetricKey(model.MetricTypeGauge, "id1")},
			want: want{
				got: model.NewGaugeMetric("id1", -1.23),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"unknown metric fails": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: false,
			},
			args: args{key: model.MetricKey{Type: model.MetricType("something"), ID: "id1"}},
			want: want{
				got: model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrUnsupportedMetricType)
				},
			},
		},
		"database timeout results in error": {
			timeout: 10 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      25 * time.Millisecond,
				mockQuery:      `SELECT value FROM metrics_counter WHERE metric_id = $1`,
				mockArgs:       []driver.Value{"id1"},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"value"})
					rows.AddRow(123)
					return rows
				},
			},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: want{
				got: model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "canceling query due to user request")
				},
			},
		},
		"database unknown failure": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        errors.New("database crashed"),
				mockDelay:      25 * time.Millisecond,
				mockQuery:      `SELECT value FROM metrics_counter WHERE metric_id = $1`,
				mockArgs:       []driver.Value{"id1"},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"value"})
					rows.AddRow(123)
					return rows
				},
			},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: want{
				got: model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "database crashed")
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fields.expectMockCall {
				mock.ExpectQuery(tt.fields.mockQuery).
					WithArgs(tt.fields.mockArgs...).
					WillDelayFor(tt.fields.mockDelay).
					WillReturnError(tt.fields.mockErr).
					WillReturnRows(tt.fields.mockRows())
			}
			s := &sqlStorage{
				db: db,
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			// Act
			got, err := s.Get(ctx, tt.args.key)

			// Assert
			tt.want.wantErr(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlStorage_GetAll(t *testing.T) {
	type fields struct {
		expectMockCall bool
		mockErr        error
		mockDelay      time.Duration
		mockQuery      string
		mockRows       func() *sqlmock.Rows
	}
	type want struct {
		got     []model.Metric
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		timeout time.Duration
		fields  fields
		want    want
	}
	tests := map[string]testcase{
		"empty storage returns empty slice without error": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows:       func() *sqlmock.Rows { return sqlmock.NewRows(nil) },
			},
			want: want{
				got: []model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"storage contains only counters": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"metric_type", "metric_id", "value"})
					rows.AddRow("counter", "id1", 123)
					rows.AddRow("counter", "id2", -123)
					rows.AddRow("counter", "id3", 456)
					return rows
				},
			},
			want: want{
				got: []model.Metric{
					model.NewCounterMetric("id3", 456),
					model.NewCounterMetric("id1", 123),
					model.NewCounterMetric("id2", -123),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"storage contains only gauges": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"metric_type", "metric_id", "value"})
					rows.AddRow("gauge", "id1", 1.23)
					rows.AddRow("gauge", "id2", -1.23)
					rows.AddRow("gauge", "id3", 456)
					rows.AddRow("gauge", "id4", -789)
					return rows
				},
			},
			want: want{
				got: []model.Metric{
					model.NewGaugeMetric("id3", 456),
					model.NewGaugeMetric("id1", 1.23),
					model.NewGaugeMetric("id4", -789),
					model.NewGaugeMetric("id2", -1.23),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"storage contains mix of counters and gauges": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"metric_type", "metric_id", "value"})
					rows.AddRow("counter", "id4", 789)
					rows.AddRow("gauge", "id1", 1.23)
					rows.AddRow("gauge", "id2", -1.23)
					rows.AddRow("counter", "id1", 123)
					rows.AddRow("gauge", "id3", 456)
					rows.AddRow("counter", "id2", -123)
					rows.AddRow("counter", "id3", 456)
					rows.AddRow("gauge", "id4", -789)
					return rows
				},
			},
			want: want{
				got: []model.Metric{
					model.NewGaugeMetric("id3", 456),
					model.NewGaugeMetric("id1", 1.23),
					model.NewGaugeMetric("id4", -789),
					model.NewGaugeMetric("id2", -1.23),
					model.NewCounterMetric("id3", 456),
					model.NewCounterMetric("id4", 789),
					model.NewCounterMetric("id1", 123),
					model.NewCounterMetric("id2", -123),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"slow storage request cancelled": {
			timeout: 10 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      25 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"metric_type", "metric_id", "value"})
					rows.AddRow("counter", "id4", 789)
					rows.AddRow("gauge", "id1", 1.23)
					rows.AddRow("gauge", "id2", -1.23)
					rows.AddRow("counter", "id1", 123)
					return rows
				},
			},
			want: want{
				got: []model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "canceling query")
				},
			},
		},
		"storage unknown error": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        errors.New("oops"),
				mockDelay:      25 * time.Millisecond,
				mockQuery:      `(SELECT 'counter' AS metric_type, metric_id, value::NUMERIC FROM metrics_counter) UNION ALL (SELECT 'gauge' AS metric_type, metric_id, value::NUMERIC FROM metrics_gauge)`,
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{"metric_type", "metric_id", "value"})
					rows.AddRow("counter", "id4", 789)
					rows.AddRow("gauge", "id1", 1.23)
					rows.AddRow("gauge", "id2", -1.23)
					rows.AddRow("counter", "id1", 123)
					return rows
				},
			},
			want: want{
				got: []model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "oops")
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fields.expectMockCall {
				mock.ExpectQuery(tt.fields.mockQuery).
					WillDelayFor(tt.fields.mockDelay).
					WillReturnError(tt.fields.mockErr).
					WillReturnRows(tt.fields.mockRows())
			}
			s := &sqlStorage{
				db: db,
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			// Act
			got, err := s.GetAll(ctx)

			// Assert
			tt.want.wantErr(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
			assert.ElementsMatch(t, tt.want.got, got)
		})
	}
}

func Test_sqlStorage_Set(t *testing.T) {
	type fields struct {
		expectMockCall bool
		mockErr        error
		mockDelay      time.Duration
		mockQuery      string
		mockArgs       []driver.Value
		mockResult     driver.Result
	}
	type args struct {
		metric model.Metric
	}
	type want struct {
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		timeout time.Duration
		fields  fields
		args    args
		want    want
	}
	tests := map[string]testcase{
		"empty metric is ignored": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: false,
			},
			args: args{metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"}},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"unknown metric returns error": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: false,
			},
			args: args{metric: func() model.Metric {
				var value float64
				m := model.Metric{Type: model.MetricType("something"), ID: "id1", Value: &value}
				return m
			}()},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrUnsupportedMetricType)
				},
			},
		},
		"counter metrics updated/inserted": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `INSERT INTO metrics_counter (metric_id, value) VALUES ($1, $2) ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`,
				mockArgs:       []driver.Value{"id1", 123},
				mockResult:     driver.RowsAffected(1),
			},
			args: args{metric: model.NewCounterMetric("id1", 123)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"gauge metrics updated/inserted": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `INSERT INTO metrics_gauge (metric_id, value) VALUES ($1, $2) ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`,
				mockArgs:       []driver.Value{"id1", -1.23},
				mockResult:     driver.RowsAffected(1),
			},
			args: args{metric: model.NewGaugeMetric("id1", -1.23)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"slow storage update cancelled": {
			timeout: 10 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      25 * time.Millisecond,
				mockQuery:      `INSERT INTO metrics_counter (metric_id, value) VALUES ($1, $2) ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`,
				mockArgs:       []driver.Value{"id1", 123},
				mockResult:     driver.RowsAffected(1),
			},
			args: args{metric: model.NewCounterMetric("id1", 123)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "canceled query")
				},
			},
		},
		"internal storage error": {
			timeout: 100 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        errors.New("database crash"),
				mockDelay:      15 * time.Millisecond,
				mockQuery:      `INSERT INTO metrics_counter (metric_id, value) VALUES ($1, $2) ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`,
				mockArgs:       []driver.Value{"id1", 123},
				mockResult:     driver.ResultNoRows,
			},
			args: args{metric: model.NewCounterMetric("id1", 123)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "database crash")
				},
			},
		},
		"no rows updated results in error": {
			timeout: 10 * time.Millisecond,
			fields: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      `INSERT INTO metrics_counter (metric_id, value) VALUES ($1, $2) ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`,
				mockArgs:       []driver.Value{"id1", 123},
				mockResult:     driver.ResultNoRows,
			},
			args: args{metric: model.NewCounterMetric("id1", 123)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.Errorf(t, err, "nothing was updated")
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fields.expectMockCall {
				mock.ExpectExec(tt.fields.mockQuery).
					WithArgs(tt.fields.mockArgs...).
					WillDelayFor(tt.fields.mockDelay).
					WillReturnError(tt.fields.mockErr).
					WillReturnResult(tt.fields.mockResult)
			}
			s := &sqlStorage{
				db: db,
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			// Act
			err = s.Set(ctx, tt.args.metric)

			// Assert
			tt.want.wantErr(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
