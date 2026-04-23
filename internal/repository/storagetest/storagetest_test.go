package storagetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/bq2cd/yp-go-metrics/internal/repository"
)

func TestNewMockStorage(t *testing.T) {
	type args struct {
		metrics []model.Metric
	}
	type want struct {
		isFaulty bool
		metrics  model.MetricSet
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default",
			args: args{},
			want: want{metrics: model.NewMetricSet()},
		},
		{
			name: "some metrics",
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{
				metrics: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
		},
		{
			name: "some metrics with duplicate ids",
			args: args{
				metrics: []model.Metric{
					model.NewCounterMetric("id1", 5),
					model.NewGaugeMetric("id2", -3.3),
					model.NewCounterMetric("id1", -5),
					model.NewGaugeMetric("id2", 4),
					model.NewCounterMetric("id3", 15),
				},
			},
			want: want{
				metrics: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", -5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", 4),
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", 15),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockStorage(tt.args.metrics...)
			assert.Equal(t, tt.want.isFaulty, s.isFaulty)
			assert.Equal(t, tt.want.metrics, s.data)
		})
	}
}

func TestMockStorage_Get(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
	}
	type args struct {
		k model.MetricKey
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty storage",
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, repository.ErrMetricNotFound)
			},
		},
		{
			name: "normal get",
			fields: fields{
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "id1")},
			want: model.NewCounterMetric("id1", 5),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "faulty get",
			fields: fields{
				isFaulty:  true,
				triggerID: "zztop",
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"):   model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeCounter, "zztop"): model.NewCounterMetric(FaultyStorageErrorTrigger, 17),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):     model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{k: model.NewMetricKey(model.MetricTypeCounter, "zztop")},
			want: model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrFaultyStorage)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			got, err := s.Get(t.Context(), tt.args.k)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMockStorage_GetAll(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
	}
	type want struct {
		metrics []model.Metric
		err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "empty storage",
			fields: fields{
				data: model.NewMetricSet(),
			},
			want: want{metrics: []model.Metric{}},
		},
		{
			name: "some metrics",
			fields: fields{
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", -8),
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{metrics: []model.Metric{
				model.NewCounterMetric("id1", 5),
				model.NewGaugeMetric("id2", -3.3),
				model.NewCounterMetric("id3", -8),
			}},
		},
		{
			name: "faulty metric",
			fields: fields{
				isFaulty:  true,
				triggerID: "tada",
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id3"): model.NewCounterMetric("id3", -8),
					model.NewMetricKey(model.MetricTypeGauge, "tada"):  model.NewGaugeMetric("tada", 1.8),
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			want: want{metrics: []model.Metric{}, err: ErrFaultyStorage},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			got, err := s.GetAll(t.Context())
			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want.metrics, got)
		})
	}
}

func TestMockStorage_Set(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
	}
	type args struct {
		m model.Metric
	}
	type want struct {
		m   model.Metric
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "empty storage",
			fields: fields{
				data: model.NewMetricSet(),
			},
			args: args{m: model.NewCounterMetric("id1", 5)},
			want: want{m: model.NewCounterMetric("id1", 5)},
		},
		{
			name: "add new",
			fields: fields{
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("id3", -5)},
			want: want{m: model.NewCounterMetric("id3", -5)},
		},
		{
			name: "overwrite existing",
			fields: fields{
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("id1", -5)},
			want: want{m: model.NewCounterMetric("id1", -5)},
		},
		{
			name: "faulty metric",
			fields: fields{
				isFaulty:  true,
				triggerID: "oops",
				data: model.MetricSet{
					model.NewMetricKey(model.MetricTypeCounter, "id1"): model.NewCounterMetric("id1", 5),
					model.NewMetricKey(model.MetricTypeGauge, "id2"):   model.NewGaugeMetric("id2", -3.3),
				},
			},
			args: args{m: model.NewCounterMetric("oops", -5)},
			want: want{m: model.Metric{}, err: ErrFaultyStorage},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}

			err := s.Set(t.Context(), tt.args.m)

			if tt.want.err != nil {
				require.ErrorIs(t, err, tt.want.err)
				return
			}

			require.NoError(t, err)
			require.Contains(t, s.data, tt.want.m.Key())
			assert.Equal(t, tt.want.m, s.data[tt.want.m.Key()])
		})
	}
}

func TestMockStorage_MakeFaulty(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
	}
	type want struct {
		isFaulty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name:   "normal -> faulty",
			fields: fields{data: model.NewMetricSet(), isFaulty: false, triggerID: "something"},
			want:   want{isFaulty: true},
		},
		{
			name:   "faulty -> faulty",
			fields: fields{data: model.NewMetricSet(), isFaulty: true, triggerID: "something"},
			want:   want{isFaulty: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			assert.Equal(t, tt.want.isFaulty, s.MakeFaulty().isFaulty)
		})
	}
}

func TestMockStorage_MakeNormal(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
	}
	type want struct {
		isFaulty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name:   "normal -> normal",
			fields: fields{data: model.NewMetricSet(), isFaulty: false, triggerID: "something"},
			want:   want{isFaulty: false},
		},
		{
			name:   "faulty -> normal",
			fields: fields{data: model.NewMetricSet(), isFaulty: true, triggerID: "something"},
			want:   want{isFaulty: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			assert.Equal(t, tt.want.isFaulty, s.MakeNormal().isFaulty)
		})
	}
}

func TestMockStorage_GetMulti(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
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
				data: model.NewMetricSet(),
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
		"non-empty storage, multiple keys requested, faulty metrics cause error and partial result": {
			fields: fields{
				isFaulty:  true,
				triggerID: "never-again",
				data: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id3", 0.5),
					model.NewCounterMetric("never-again", 17),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
				),
			},
			args: args{keys: model.NewMetricKeySet(
				model.NewMetricKey(model.MetricTypeCounter, "id5"),
				model.NewMetricKey(model.MetricTypeGauge, "id4"),
				model.NewMetricKey(model.MetricTypeCounter, "never-again"),
			)},
			want: want{
				got: []model.Metric{
					model.NewCounterMetric("id5", -15),
				},
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrFaultyStorage)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			got, err := s.GetMulti(t.Context(), tt.args.keys)
			tt.want.wantErr(t, err)
			assert.ElementsMatch(t, tt.want.got, got)
		})
	}
}

func TestMockStorage_SetMulti(t *testing.T) {
	type fields struct {
		data      model.MetricSet
		isFaulty  bool
		triggerID string
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
		"non-empty storage, multiple metrics, faulty metrics partial update": {
			fields: fields{
				isFaulty:  true,
				triggerID: "never-again",
				data: model.NewMetricSet(
					model.NewCounterMetric("id2", 15),
					model.NewGaugeMetric("id4", 1.5),
					model.NewCounterMetric("id7", -18),
				),
			},
			args: args{metrics: model.NewMetricSet(
				model.NewCounterMetric("id1", 5),
				model.NewCounterMetric("id2", -15),
				model.NewGaugeMetric("never-again", 0.5),
				model.NewGaugeMetric("id4", -1.5),
				model.NewCounterMetric("id5", -15),
				model.NewGaugeMetric("id6", -0.5),
			)},
			want: want{
				got: model.NewMetricSet(
					model.NewCounterMetric("id1", 5),
					model.NewCounterMetric("id2", -15),
					model.NewGaugeMetric("id4", -1.5),
					model.NewCounterMetric("id5", -15),
					model.NewGaugeMetric("id6", -0.5),
					model.NewCounterMetric("id7", -18),
				),
				wantErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, ErrFaultyStorage)
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &MockStorage{
				data:      tt.fields.data,
				isFaulty:  tt.fields.isFaulty,
				triggerID: tt.fields.triggerID,
			}
			err := s.SetMulti(t.Context(), tt.args.metrics)
			tt.want.wantErr(t, err)
			assert.Equal(t, tt.want.got, s.data)
		})
	}
}
