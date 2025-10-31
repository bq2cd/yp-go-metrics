package repository

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *memStorage
	}{
		{
			name: "empty storage",
			want: &memStorage{data: model.NewMetricSet()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMemStorage())
		})
	}
}

func Test_memStorage_Get(t *testing.T) {
	type fields struct {
		data model.MetricSet
	}
	type args struct {
		key model.MetricKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{data: model.NewMetricSet()},
			args:   args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMetricNotFound)
			},
		},
		{
			name: "metric exists",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): func() model.Metric {
					var value int64 = 9
					return model.Metric{
						ID:    "id1",
						Type:  model.MetricTypeCounter,
						Delta: &value,
						Hash:  "counter1",
					}
				}(),
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				}}},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: func() model.Metric {
				var value int64 = 9
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}(),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "empty metric requested",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				}}},
			args: args{key: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMetricNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			got, err := s.Get(t.Context(), tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_memStorage_Set(t *testing.T) {
	type fields struct {
		data model.MetricSet
	}
	type args struct {
		metric func() model.Metric
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
		contains  func(assert.TestingT, model.MetricSet, model.Metric)
	}{
		{
			name:   "empty storage",
			fields: fields{data: model.NewMetricSet()},
			args: args{metric: func() model.Metric {
				var value int64 = 10
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data model.MetricSet, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
		{
			name:   "empty metric not added",
			fields: fields{data: model.NewMetricSet()},
			args: args{metric: func() model.Metric {
				var value int64 = 5
				return model.Metric{
					ID:    "id3",
					Type:  model.MetricTypeGauge,
					Delta: &value,
					Hash:  "gauge3",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data model.MetricSet, metric model.Metric) {
				assert.NotContains(t, data, metric.Key())
			},
		},
		{
			name: "metric added",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 5
				return model.Metric{
					ID:    "id3",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter3",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data model.MetricSet, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
		{
			name: "metric replaced",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeCounter,
					Hash: "counter1",
				},
				model.NewMetricKey(model.MetricTypeGauge, "id1"): model.Metric{
					ID:   "id1",
					Type: model.MetricTypeGauge,
					Hash: "gauge1",
				},
			}},
			args: args{metric: func() model.Metric {
				var value int64 = 15
				return model.Metric{
					ID:    "id1",
					Type:  model.MetricTypeCounter,
					Delta: &value,
					Hash:  "counter1",
				}
			}},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
			contains: func(t assert.TestingT, data model.MetricSet, metric model.Metric) {
				assert.Contains(t, data, metric.Key())
				assert.Equal(t, metric, data[metric.Key()])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			metric := tt.args.metric()
			tt.assertion(t, s.Set(t.Context(), metric))
			tt.contains(t, tt.fields.data, metric)
		})
	}
}

func Test_memStorage_GetAll(t *testing.T) {
	type fields struct {
		data model.MetricSet
	}
	tests := []struct {
		name      string
		fields    fields
		want      []model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty storage",
			fields: fields{data: model.NewMetricSet()},
			want:   []model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "multiple metrics",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 10),
				model.NewMetricKey(model.MetricTypeGauge, "id1"):   model.NewGaugeMetric("id2", 1.5),
			}},
			want: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id2", 1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "storage with empty metrics",
			fields: fields{data: model.MetricSet{
				model.NewMetricKey(model.MetricTypeCounter, "id1"):   model.NewCounterMetric("id1", 10),
				model.NewMetricKey(model.MetricTypeGauge, "id1"):     model.NewGaugeMetric("id2", 1.5),
				model.NewMetricKey(model.MetricTypeCounter, ""):      model.Metric{Type: model.MetricTypeCounter},
				model.NewMetricKey("", ""):                           model.Metric{},
				model.NewMetricKey(model.MetricTypeGauge, "emptyMe"): model.Metric{ID: "emptyMe", Type: model.MetricTypeGauge},
			}},
			want: []model.Metric{model.NewCounterMetric("id1", 10), model.NewGaugeMetric("id2", 1.5)},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			got, err := s.GetAll(t.Context())
			tt.assertion(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_memStorage_GetMulti(t *testing.T) {
	type fields struct {
		data model.MetricSet
	}
	type args struct {
		keys model.MetricKeySet
	}
	type want struct {
		got     []model.Metric
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty storage, empty keys requested": {
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{keys: model.NewMetricKeySet()},
			want: want{
				got: []model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"empty storage, multiple keys requested": {
			fields: fields{
				data: map[model.MetricKey]model.Metric{},
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id1"),
				model.NewMetricKey(model.MetricTypeCounter, "id2"),
			)},
			want: want{
				got: []model.Metric{},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"non-empty storage, single key requested": {
			fields: fields{
				data: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewGaugeMetric("id4", 1.5),
				),
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id2"),
			)},
			want: want{
				got: []model.Metric{
					model.NewCounterMetric("id2", 15),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"non-empty storage, multiple keys requested": {
			fields: fields{
				data: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
				),
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id5"),
				model.NewMetricKey(model.MetricTypeGauge, "id4"),
			)},
			want: want{
				got: []model.Metric{
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id5", -15),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"non-empty storage, multiple keys requested, non-existent metrics skipped": {
			fields: fields{
				data: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
				),
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id5"),
				model.NewMetricKey(model.MetricTypeGauge, "id4"),
				model.NewMetricKey(model.MetricTypeCounter, "id3"),
				model.NewMetricKey(model.MetricTypeGauge, "id2"),
			)},
			want: want{
				got: []model.Metric{
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id5", -15),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"non-empty storage, multiple keys requested, empty metrics skipped": {
			fields: fields{
				data: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.Metric{Type: model.MetricTypeGauge, ID: "id4"},
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
				),
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id5"),
				model.NewMetricKey(model.MetricTypeGauge, "id4"),
			)},
			want: want{
				got: []model.Metric{
					model.NewCounterMetric("id5", -15),
				},
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			got, err := s.GetMulti(t.Context(), tt.args.keys)
			tt.want.wantErr(t, err)
			assert.ElementsMatch(t, tt.want.got, got)
		})
	}
}

func Test_memStorage_SetMulti(t *testing.T) {
	type fields struct {
		data model.MetricSet
	}
	type args struct {
		metrics model.MetricSet
	}
	type want struct {
		got     model.MetricSet
		wantErr func(testing.TB, error)
	}
	type testcase struct {
		fields fields
		args   args
		want   want
	}
	tests := map[string]testcase{
		"empty storage, empty metrics, nothing happens": {
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{metrics: model.NewMetricSet()},
			want: want{
				got: model.NewMetricSet(),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"empty storage, single metric": {
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
			)},
			want: want{
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
				),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"empty storage, multiple metrics": {
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
				model.NewCounterMetric("id2", 15),
				model.NewGaugeMetric("id3", 0.5),
				model.NewGaugeMetric("id4", 1.5),
				model.NewCounterMetric("id5", -15),
				model.NewGaugeMetric("id6", -0.5),
			)},
			want: want{
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
				),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
		"non-empty storage, multiple metrics, existing overwritten": {
			fields: fields{
				data: model.NewMetricSet(
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id7", -18),
				),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
				model.NewCounterMetric("id2", -15),
				model.NewGaugeMetric("id3", 0.5),
				model.NewGaugeMetric("id4", -1.5),
				model.NewCounterMetric("id5", -15),
				model.NewGaugeMetric("id6", -0.5),
			)},
			want: want{
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewGaugeMetric("id4", -1.5),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
					model.NewCounterMetric("id7", -18),
				),
				wantErr: func(t testing.TB, err error) {
					require.NoError(t, err)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &memStorage{
				data: tt.fields.data,
			}
			err := s.SetMulti(t.Context(), tt.args.metrics)
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.got, s.data)
		})
	}
}
