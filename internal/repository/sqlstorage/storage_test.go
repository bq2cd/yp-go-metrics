package sqlstorage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	dbconfig "github.com/bq2cd/yp-go-metrics/internal/config/db"
	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest/mockretrierfactory"
)

const (
	databaseDriver = "pgx"
)

func TestNew(t *testing.T) {
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
					assert.NotNil(t, got.retrierFactory)
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
					assert.NotNil(t, got.retrierFactory)
				},
			},
		}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)

			cfg, err := dbconfig.New(tt.args.dbURL)

			tt.want.assertConfig(t, &cfg, err)
			got, err := New(cfg, retrierFactory)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)

			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectPing().WillDelayFor(tt.fields.mockDelay).WillReturnError(tt.fields.mockErr)
			s := &sqlStorage{
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectClose().WillReturnError(tt.fields.mockErr)
			s := &sqlStorage{
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
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

func getExpectedSelectQuery(metricType model.MetricType, numArgs int) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "SELECT metric_id, value FROM %s%s", tableNamePrefix, metricType)
	if numArgs == 0 {
		return sb.String()
	}
	sb.WriteString(` WHERE metric_id IN (`)
	for i := 1; i <= numArgs; i++ {
		fmt.Fprintf(sb, "$%d", i)
		if i < numArgs {
			sb.WriteRune(',')
		}
	}
	sb.WriteString(`)`)
	return sb.String()
}

func getExpectedInsertQuery(metricType model.MetricType, numArgs int) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "INSERT INTO %s%s (metric_id, value) VALUES ", tableNamePrefix, metricType)
	for i := 1; i <= numArgs*2; i += 2 {
		fmt.Fprintf(sb, "($%d, $%d)", i, i+1)
		if i < numArgs*2-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(" ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value")
	return sb.String()
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
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 1),
				mockArgs:       []driver.Value{`id1`},
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
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 1),
				mockArgs:       []driver.Value{`id1`},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 123)
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
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 1),
				mockArgs:       []driver.Value{`id1`},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, -1.23)
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
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 1),
				mockArgs:       []driver.Value{`id1`},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 123)
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
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 1),
				mockArgs:       []driver.Value{`id1`},
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 123)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

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
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
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
		timeout       time.Duration
		fieldsCounter fields
		fieldsGauge   fields
		want          want
	}
	tests := map[string]testcase{
		"empty storage returns empty slice without error": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows:       func() *sqlmock.Rows { return sqlmock.NewRows(nil) },
			},
			fieldsGauge: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
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
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 123)
					rows.AddRow(`id2`, -123)
					rows.AddRow(`id3`, 456)
					return rows
				},
			},
			fieldsGauge: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
				mockRows:       func() *sqlmock.Rows { return sqlmock.NewRows(nil) },
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
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        sql.ErrNoRows,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows:       func() *sqlmock.Rows { return sqlmock.NewRows(nil) },
			},
			fieldsGauge: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 1.23)
					rows.AddRow(`id2`, -1.23)
					rows.AddRow(`id3`, 456)
					rows.AddRow(`id4`, -789)
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
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id4`, 789)
					rows.AddRow(`id1`, 123)
					rows.AddRow(`id2`, -123)
					rows.AddRow(`id3`, 456)
					return rows
				},
			},
			fieldsGauge: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 1.23)
					rows.AddRow(`id2`, -1.23)
					rows.AddRow(`id3`, 456)
					rows.AddRow(`id4`, -789)
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
			timeout: 15 * time.Millisecond,
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      10 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id4`, 789)
					rows.AddRow(`id1`, 123)
					return rows
				},
			},
			fieldsGauge: fields{
				expectMockCall: true,
				mockErr:        nil,
				mockDelay:      10 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 1.23)
					rows.AddRow(`id2`, -1.23)
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
			fieldsCounter: fields{
				expectMockCall: true,
				mockErr:        errors.New("oops"),
				mockDelay:      25 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeCounter, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id4`, 789)
					rows.AddRow(`id1`, 123)
					return rows
				},
			},
			fieldsGauge: fields{
				// this will not get called since we return on the first error and metric types are sorted (counter comes first)
				expectMockCall: false,
				mockErr:        errors.New("oops"),
				mockDelay:      25 * time.Millisecond,
				mockQuery:      getExpectedSelectQuery(model.MetricTypeGauge, 0),
				mockRows: func() *sqlmock.Rows {
					rows := sqlmock.NewRows([]string{`metric_id`, `value`})
					rows.AddRow(`id1`, 1.23)
					rows.AddRow(`id2`, -1.23)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fieldsCounter.expectMockCall {
				mock.ExpectQuery(tt.fieldsCounter.mockQuery).
					WillDelayFor(tt.fieldsCounter.mockDelay).
					WillReturnError(tt.fieldsCounter.mockErr).
					WillReturnRows(tt.fieldsCounter.mockRows())
			}
			if tt.fieldsGauge.expectMockCall {
				mock.ExpectQuery(tt.fieldsGauge.mockQuery).
					WillDelayFor(tt.fieldsGauge.mockDelay).
					WillReturnError(tt.fieldsGauge.mockErr).
					WillReturnRows(tt.fieldsGauge.mockRows())
			}

			s := &sqlStorage{
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
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
		expectTxBegin    bool
		expectTxCommit   bool
		expectTxRollback bool
		expectQuery      bool
		mockErr          error
		mockDelay        time.Duration
		mockQuery        string
		mockArgs         []driver.Value
		mockResult       driver.Result
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
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
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
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
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
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 1),
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
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeGauge, 1),
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
				expectTxBegin:  true,
				expectTxCommit: false,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      25 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 1),
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
				expectTxBegin:  true,
				expectTxCommit: false,
				expectQuery:    true,
				mockErr:        errors.New("database crash"),
				mockDelay:      15 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 1),
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
		"no rows updated should not fail": {
			timeout: 10 * time.Millisecond,
			fields: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 1),
				mockArgs:       []driver.Value{"id1", 123},
				mockResult:     driver.ResultNoRows,
			},
			args: args{metric: model.NewCounterMetric("id1", 123)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fields.expectTxBegin {
				mock.ExpectBegin()
			}
			if tt.fields.expectQuery {
				mock.ExpectExec(tt.fields.mockQuery).
					WithArgs(tt.fields.mockArgs...).
					WillDelayFor(tt.fields.mockDelay).
					WillReturnError(tt.fields.mockErr).
					WillReturnResult(tt.fields.mockResult)
			}
			if tt.fields.expectTxRollback {
				mock.ExpectRollback()
			}
			if tt.fields.expectTxCommit {
				mock.ExpectCommit()
			}
			s := &sqlStorage{
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
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

func Test_sqlStorage_SetMulti(t *testing.T) {
	type fields struct {
		expectTxBegin    bool
		expectTxCommit   bool
		expectTxRollback bool
		expectQuery      bool
		mockErr          error
		mockDelay        time.Duration
		mockQuery        string
		mockArgs         []driver.Value
		mockResult       driver.Result
	}
	type args struct {
		metrics model.MetricSet
	}
	type want struct {
		wantErr    func(testing.TB, error)
		numRetries int
	}
	type testcase struct {
		timeout       time.Duration
		fieldsCounter fields
		fieldsGauge   fields
		args          args
		want          want
	}
	tests := map[string]testcase{
		"empty metric set does not query database": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			fieldsGauge: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			args: args{metrics: model.NewMetricSet()},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"empty metrics never reach database": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			fieldsGauge: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			args: args{metrics: model.NewMetricSet(
				model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
				model.Metric{Type: model.MetricTypeGauge, ID: "id2"},
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"unknown metrics result in error": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			fieldsGauge: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			args: args{metrics: model.NewMetricSet(
				func() model.Metric {
					var value float64
					m := model.Metric{Type: model.MetricType("something"), ID: "id1", Value: &value}
					return m
				}(),
				model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
				model.Metric{Type: model.MetricTypeGauge, ID: "id2"},
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrUnsupportedMetricType)
				},
			},
		},
		"only counters are updated/inserted": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 3),
				mockArgs:       []driver.Value{"id1", 123, "id2", -123, "id3", 456},
				mockResult:     driver.RowsAffected(2),
			},
			fieldsGauge: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewCounterMetric("id3", 456),
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"only gauges are updated/inserted": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			fieldsGauge: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeGauge, 4),
				mockArgs:       []driver.Value{"id1", 1.23, "id2", -12.3, "id3", float64(456), "id4", float64(-789)},
				mockResult:     driver.RowsAffected(4),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewGaugeMetric("id1", 1.23),
				model.NewGaugeMetric("id2", -12.3),
				model.NewGaugeMetric("id3", 456),
				model.NewGaugeMetric("id4", -789),
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"both counter and gauges are updated/inserted": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 3),
				mockArgs:       []driver.Value{"id5", 123, "id6", -123, "id7", 456},
				mockResult:     driver.RowsAffected(3),
			},
			fieldsGauge: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        nil,
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeGauge, 4),
				mockArgs:       []driver.Value{"id1", 1.23, "id2", -12.3, "id3", float64(456), "id4", float64(-789)},
				mockResult:     driver.RowsAffected(4),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewGaugeMetric("id2", -12.3),
				model.NewGaugeMetric("id1", 1.23),
				model.Metric{Type: model.MetricTypeCounter, ID: "id9"},
				model.NewGaugeMetric("id3", 456),
				model.NewCounterMetric("id6", -123),
				model.NewGaugeMetric("id4", -789),
				model.NewCounterMetric("id5", 123),
				model.Metric{Type: model.MetricTypeGauge, ID: "id8"},
				model.NewCounterMetric("id7", 456),
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"database returns retryable error on update/insert": {
			timeout: 100 * time.Millisecond,
			fieldsCounter: fields{
				expectTxBegin:  true,
				expectTxCommit: true,
				expectQuery:    true,
				mockErr:        &pgconn.PgError{Code: pgerrcode.ConnectionFailure, Message: "oops"},
				mockDelay:      5 * time.Millisecond,
				mockQuery:      getExpectedInsertQuery(model.MetricTypeCounter, 3),
				mockArgs:       []driver.Value{"id1", 123, "id2", -123, "id3", 456},
				mockResult:     driver.RowsAffected(2),
			},
			fieldsGauge: fields{
				expectTxBegin:  false,
				expectTxCommit: false,
				expectQuery:    false,
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 123),
				model.NewCounterMetric("id2", -123),
				model.NewCounterMetric("id3", 456),
			)},
			want: want{
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
				numRetries: 2,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

			if tt.want.numRetries > 0 {
				calls := []any{}
				for range tt.want.numRetries {
					calls = append(calls, retrierFactory.Strategy.EXPECT().NextDelay().Return(1*time.Millisecond, true))
					calls = append(calls, retrierFactory.Sleeper.EXPECT().Sleep(gomock.Any(), 1*time.Millisecond).Return(nil))
				}
				gomock.InOrder(calls...)
			}

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			if tt.fieldsCounter.expectTxBegin {
				mock.ExpectBegin()
			}
			if tt.fieldsCounter.expectQuery {
				mock.ExpectExec(tt.fieldsCounter.mockQuery).
					WithArgs(tt.fieldsCounter.mockArgs...).
					WillDelayFor(tt.fieldsCounter.mockDelay).
					WillReturnError(tt.fieldsCounter.mockErr).
					WillReturnResult(tt.fieldsCounter.mockResult)
			}
			for i := range tt.want.numRetries {
				var errFinal error
				if i < tt.want.numRetries-1 {
					errFinal = tt.fieldsCounter.mockErr
				} else {
					errFinal = nil
				}
				mock.ExpectBegin()
				mock.ExpectExec(tt.fieldsCounter.mockQuery).
					WithArgs(tt.fieldsCounter.mockArgs...).
					WillDelayFor(tt.fieldsCounter.mockDelay).
					WillReturnError(errFinal).
					WillReturnResult(tt.fieldsCounter.mockResult)
			}
			if tt.fieldsCounter.expectTxCommit {
				mock.ExpectCommit()
			}

			if tt.fieldsGauge.expectTxBegin {
				mock.ExpectBegin()
			}
			if tt.fieldsGauge.expectQuery {
				mock.ExpectExec(tt.fieldsGauge.mockQuery).
					WithArgs(tt.fieldsGauge.mockArgs...).
					WillDelayFor(tt.fieldsGauge.mockDelay).
					WillReturnError(tt.fieldsGauge.mockErr).
					WillReturnResult(tt.fieldsGauge.mockResult)
			}
			if tt.fieldsGauge.expectTxCommit {
				mock.ExpectCommit()
			}

			s := &sqlStorage{
				db:             sqlx.NewDb(db, databaseDriver),
				retrierFactory: retrierFactory,
			}

			ctx, cancel := context.WithTimeout(t.Context(), tt.timeout)
			defer cancel()

			// Act
			err = s.SetMulti(ctx, tt.args.metrics)

			// Assert
			tt.want.wantErr(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_sqlStorage_GetMulti(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		ctx  context.Context
		keys model.MetricKeySet
	}
	type want struct {
		got []model.Metric
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			retrierFactory := mockretrierfactory.NewMockRetrierFactory(ctrl)
			retrierFactory.Strategy.EXPECT().Name().Return("mock_strategy")

			s := &sqlStorage{
				db:             tt.fields.db,
				retrierFactory: retrierFactory,
			}
			got, err := s.GetMulti(tt.args.ctx, tt.args.keys)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
