package service

import (
	"errors"
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWriteCloser struct {
	mock.Mock
	writeErr error
	closeErr error
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	m.Called(p)
	return 0, m.writeErr
}
func (m *mockWriteCloser) Close() error {
	m.Called()
	return m.closeErr
}

type mockReadCloser struct {
	mock.Mock
	readErr  error
	closeErr error
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	m.Called(p)
	return 0, m.readErr
}
func (m *mockReadCloser) Close() error {
	m.Called()
	return m.closeErr
}

func TestNewMetricSnapshotter(t *testing.T) {
	type args struct {
		storer  MetricStorer
		encoder MetricEncoder
		decoder MetricDecoder
	}
	tests := []struct {
		name      string
		args      args
		assertion func(*testing.T, args, *metricSnapshotter)
	}{
		{
			name: "default",
			args: args{
				storer:  &mockMetricStorer{},
				encoder: &mockMetricEncoder{},
				decoder: &mockMetricDecoder{},
			},
			assertion: func(t *testing.T, args args, got *metricSnapshotter) {
				assert.Equal(t, args.storer, got.MetricStorer)
				assert.Equal(t, args.encoder, got.encoder)
				assert.Equal(t, args.decoder, got.decoder)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.args, NewMetricSnapshotter(tt.args.storer, tt.args.encoder, tt.args.decoder))
		})
	}
}

func Test_metricSnapshotter_markDirty(t *testing.T) {
	type args struct {
		numWrites int
	}
	type want struct {
		dirtyWrites int64
	}
	tests := []struct {
		name  string
		args  args
		calls int
		want  want
	}{
		{
			name:  "single call does not block",
			args:  args{numWrites: 2},
			calls: 1,
			want: want{
				dirtyWrites: 2,
			},
		},
		{
			name:  "multiple calls do not block",
			args:  args{numWrites: 2},
			calls: 5,
			want: want{
				dirtyWrites: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				notifyCh: make(chan struct{}, 1),
			}
			for range tt.calls {
				p.markDirty(tt.args.numWrites)
			}
			assert.Equal(t, tt.want.dirtyWrites, p.dirtyWrites.Load())
			assert.Equal(t, struct{}{}, <-p.notifyCh)
			select {
			case <-p.notifyCh:
				assert.Truef(t, false, "channel must be empty")
			default:
				assert.True(t, true)
			}
		})
	}
}

func Test_metricSnapshotter_StoreSingle(t *testing.T) {
	type fields struct {
		MetricStorer *mockMetricStorer
	}
	type args struct {
		m model.Metric
	}
	type want struct {
		dirtyWrites  int64
		channelEmpty bool
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no error",
			fields: fields{
				MetricStorer: &mockMetricStorer{},
			},
			args: args{m: model.NewCounterMetric("id1", 123)},
			want: want{
				dirtyWrites:  1,
				channelEmpty: false,
			},
			assertion: assert.NoError,
		},
		{
			name: "storer error",
			fields: fields{
				MetricStorer: &mockMetricStorer{err: errors.New("oops")},
			},
			args: args{m: model.NewCounterMetric("id1", 123)},
			want: want{
				dirtyWrites:  0,
				channelEmpty: true,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				notifyCh:     make(chan struct{}, 1),
			}
			tt.fields.MetricStorer.On("StoreSingle", tt.args.m).Return(mock.AnythingOfType("error")).Once()

			err := p.StoreSingle(tt.args.m)

			tt.assertion(t, err)
			tt.fields.MetricStorer.AssertExpectations(t)
			assert.Equal(t, tt.want.dirtyWrites, p.dirtyWrites.Load())
			select {
			case v := <-p.notifyCh:
				assert.Equal(t, struct{}{}, v)
			default:
				assert.True(t, tt.want.channelEmpty)
			}
		})
	}
}

func Test_metricSnapshotter_StoreBatch(t *testing.T) {
	type fields struct {
		MetricStorer *mockMetricStorer
	}
	type args struct {
		metrics []model.Metric
	}
	type want struct {
		numCalls     int
		dirtyWrites  int64
		channelEmpty bool
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      want
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty metrics",
			fields: fields{
				MetricStorer: &mockMetricStorer{},
			},
			args: args{metrics: []model.Metric{}},
			want: want{
				numCalls:     0,
				dirtyWrites:  0,
				channelEmpty: true,
			},
			assertion: assert.NoError,
		},
		{
			name: "single metric",
			fields: fields{
				MetricStorer: &mockMetricStorer{},
			},
			args: args{metrics: []model.Metric{
				model.NewCounterMetric("id1", 123),
			}},
			want: want{
				numCalls:     1,
				dirtyWrites:  1,
				channelEmpty: false,
			},
			assertion: assert.NoError,
		},
		{
			name: "mutliple metrics",
			fields: fields{
				MetricStorer: &mockMetricStorer{},
			},
			args: args{metrics: []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewGaugeMetric("id2", 1.23),
				model.NewCounterMetric("id3", -456),
				model.NewGaugeMetric("id4", -4.56),
			}},
			want: want{
				numCalls:     1,
				dirtyWrites:  4,
				channelEmpty: false,
			},
			assertion: assert.NoError,
		},
		{
			name: "storer error",
			fields: fields{
				MetricStorer: &mockMetricStorer{err: errors.New("oops")},
			},
			args: args{metrics: []model.Metric{
				model.NewCounterMetric("id1", 123),
				model.NewGaugeMetric("id2", 1.23),
			}},
			want: want{
				numCalls:     1,
				dirtyWrites:  0,
				channelEmpty: true,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				MetricStorer: tt.fields.MetricStorer,
				notifyCh:     make(chan struct{}, 1),
			}
			mc := tt.fields.MetricStorer.On("StoreBatch", tt.args.metrics).Return(mock.AnythingOfType("error")).Times(tt.want.numCalls)
			if tt.want.numCalls == 0 {
				mc.Maybe()
			}

			err := p.StoreBatch(tt.args.metrics)

			tt.assertion(t, err)
			tt.fields.MetricStorer.AssertExpectations(t)
			assert.Equal(t, tt.want.dirtyWrites, p.dirtyWrites.Load())
			select {
			case v := <-p.notifyCh:
				assert.False(t, tt.want.channelEmpty)
				assert.Equal(t, struct{}{}, v)
			default:
				assert.True(t, tt.want.channelEmpty)
			}
		})
	}
}

func Test_metricSnapshotter_DumpClose(t *testing.T) {
	type fields struct {
		MetricStorer *mockMetricStorer
		encoder      *mockMetricEncoder
	}
	type args struct {
		w           *mockWriteCloser
		metrics     []model.Metric
		dirtyWrites int64
	}
	type want struct {
		dirtyWrites int64
	}
	type calls struct {
		RetrieveAll int
		EncodeBatch int
		Close       int
	}
	type errs struct {
		RetrieveAll error
		EncodeBatch error
		Close       error
	}
	type innerTest struct {
		name      string
		fields    fields
		args      args
		calls     calls
		errs      errs
		want      want
		assertion assert.ErrorAssertionFunc
	}
	type effects struct {
		RetrieveAll func(*metricSnapshotter)
		EncodeBatch func(*metricSnapshotter)
		Close       func(*metricSnapshotter)
	}
	tests := []struct {
		name    string
		effects effects
		cases   []innerTest
	}{
		{
			name: "no concurrent writes, happy path",
			effects: effects{
				RetrieveAll: func(p *metricSnapshotter) {},
				EncodeBatch: func(p *metricSnapshotter) {},
				Close:       func(p *metricSnapshotter) {},
			},
			cases: []innerTest{
				{
					name: "no writes happened",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 0,
					},
					calls: calls{
						RetrieveAll: 0,
						EncodeBatch: 0,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.NoError,
				},
				{
					name: "some writes happened",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 4.56),
							model.NewCounterMetric("id3", 456),
						},
						dirtyWrites: 2,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.NoError,
				},
			},
		},
		{
			name: "concurrent writes during retrieve",
			effects: effects{
				RetrieveAll: func(p *metricSnapshotter) { p.markDirty(2) },
				EncodeBatch: func(p *metricSnapshotter) {},
				Close:       func(p *metricSnapshotter) {},
			},
			cases: []innerTest{
				{
					name: "happy path",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 2,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 4,
					},
					assertion: assert.NoError,
				},
			},
		},
		{
			name: "concurrent writes during encode",
			effects: effects{
				RetrieveAll: func(p *metricSnapshotter) {},
				EncodeBatch: func(p *metricSnapshotter) { p.markDirty(3) },
				Close:       func(p *metricSnapshotter) {},
			},
			cases: []innerTest{
				{
					name: "happy path",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 2,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 5,
					},
					assertion: assert.NoError,
				},
			},
		},
		{
			name: "concurrent writes during close",
			effects: effects{
				RetrieveAll: func(p *metricSnapshotter) {},
				EncodeBatch: func(p *metricSnapshotter) {},
				Close:       func(p *metricSnapshotter) { p.markDirty(4) },
			},
			cases: []innerTest{
				{
					name: "happy path",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 2,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 4,
					},
					assertion: assert.NoError,
				},
			},
		},
		{
			name: "no concurrent writes, unhappy path",
			effects: effects{
				RetrieveAll: func(p *metricSnapshotter) {},
				EncodeBatch: func(p *metricSnapshotter) {},
				Close:       func(p *metricSnapshotter) {},
			},
			cases: []innerTest{
				{
					name: "no writes happened, close error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{closeErr: errors.New("oops")},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 0,
					},
					calls: calls{
						RetrieveAll: 0,
						EncodeBatch: 0,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: nil,
						Close:       errors.New("oops"),
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.Error,
				},
				{
					name: "some writes happened, retrieve error",
					fields: fields{
						MetricStorer: &mockMetricStorer{err: errors.New("oops")},
						encoder:      &mockMetricEncoder{},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 1,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 0,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: errors.New("oops"),
						EncodeBatch: nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 1,
					},
					assertion: assert.Error,
				},
				{
					name: "some writes happened, encode error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{err: errors.New("oops")},
					},
					args: args{
						w: &mockWriteCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 1,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: errors.New("oops"),
						Close:       nil,
					},
					want: want{
						dirtyWrites: 1,
					},
					assertion: assert.Error,
				},
				{
					name: "some writes happened, encode + close error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						encoder:      &mockMetricEncoder{err: errors.New("oops")},
					},
					args: args{
						w: &mockWriteCloser{closeErr: errors.New("close error")},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
						dirtyWrites: 1,
					},
					calls: calls{
						RetrieveAll: 1,
						EncodeBatch: 1,
						Close:       1,
					},
					errs: errs{
						RetrieveAll: nil,
						EncodeBatch: errors.New("oops"),
						Close:       errors.New("close error"),
					},
					want: want{
						dirtyWrites: 1,
					},
					assertion: assert.Error,
				},
			},
		},
	}
	for _, outer := range tests {
		t.Run(outer.name, func(t *testing.T) {
			for _, tt := range outer.cases {
				t.Run(tt.name, func(t *testing.T) {
					p := &metricSnapshotter{
						MetricStorer: tt.fields.MetricStorer,
						encoder:      tt.fields.encoder,
						notifyCh:     make(chan struct{}, 1),
					}
					tt.fields.MetricStorer.metrics = tt.args.metrics
					p.dirtyWrites.Store(tt.args.dirtyWrites)
					{
						mc := tt.fields.MetricStorer.On("RetrieveAll").Return(tt.args.metrics, tt.errs.RetrieveAll).Times(tt.calls.RetrieveAll)
						if tt.calls.RetrieveAll == 0 {
							mc.Maybe()
						}
						mc.Run(func(_ mock.Arguments) {
							outer.effects.RetrieveAll(p)
						})
					}
					{
						mc := tt.fields.encoder.On("EncodeBatch", tt.args.w, tt.args.metrics).Return(tt.errs.EncodeBatch).Times(tt.calls.EncodeBatch)
						if tt.calls.EncodeBatch == 0 {
							mc.Maybe()
						}
						mc.Run(func(_ mock.Arguments) {
							outer.effects.EncodeBatch(p)
						})
					}
					{
						mc := tt.args.w.On("Close").Return(tt.errs.Close).Times(tt.calls.Close)
						if tt.calls.Close == 0 {
							mc.Maybe()
						}
						mc.Run(func(_ mock.Arguments) {
							outer.effects.Close(p)
						})
					}

					err := p.DumpClose(tt.args.w)

					tt.assertion(t, err)
					tt.fields.MetricStorer.AssertExpectations(t)
					tt.fields.encoder.AssertExpectations(t)
					tt.args.w.AssertExpectations(t)
					assert.Equal(t, tt.want.dirtyWrites, p.dirtyWrites.Load())
				})
			}
		})
	}
}

func Test_metricSnapshotter_LoadClose(t *testing.T) {
	type fields struct {
		MetricStorer *mockMetricStorer
		decoder      *mockMetricDecoder
	}
	type args struct {
		r       *mockReadCloser
		metrics []model.Metric
	}
	type want struct {
		dirtyWrites int64
	}
	type calls struct {
		DecodeBatch int
		StoreBatch  int
		Close       int
	}
	type errs struct {
		DecodeBatch error
		StoreBatch  error
		Close       error
	}
	type innerTest struct {
		name      string
		fields    fields
		args      args
		calls     calls
		errs      errs
		want      want
		assertion assert.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		cases []innerTest
	}{
		{
			name: "happy path",
			cases: []innerTest{
				{
					name: "no metrics",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						decoder:      &mockMetricDecoder{},
					},
					args: args{
						r:       &mockReadCloser{},
						metrics: []model.Metric{},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  1,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: nil,
						StoreBatch:  nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.NoError,
				},
				{
					name: "some metrics",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						decoder:      &mockMetricDecoder{},
					},
					args: args{
						r: &mockReadCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
							model.NewCounterMetric("id3", -456),
							model.NewGaugeMetric("id4", -4.56),
						},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  1,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: nil,
						StoreBatch:  nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.NoError,
				},
			},
		},
		{
			name: "unhappy path",
			cases: []innerTest{
				{
					name: "decode error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						decoder:      &mockMetricDecoder{err: errors.New("oops")},
					},
					args: args{
						r:       &mockReadCloser{},
						metrics: []model.Metric{},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  0,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: errors.New("oops"),
						StoreBatch:  nil,
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.Error,
				},
				{
					name: "store error",
					fields: fields{
						MetricStorer: &mockMetricStorer{err: errors.New("oops")},
						decoder:      &mockMetricDecoder{},
					},
					args: args{
						r: &mockReadCloser{},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  1,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: nil,
						StoreBatch:  errors.New("oops"),
						Close:       nil,
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.Error,
				},
				{
					name: "close error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						decoder:      &mockMetricDecoder{},
					},
					args: args{
						r: &mockReadCloser{closeErr: errors.New("oops")},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  1,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: nil,
						StoreBatch:  nil,
						Close:       errors.New("oops"),
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.Error,
				},
				{
					name: "decode + close error",
					fields: fields{
						MetricStorer: &mockMetricStorer{},
						decoder:      &mockMetricDecoder{err: errors.New("decode failed")},
					},
					args: args{
						r: &mockReadCloser{closeErr: errors.New("oops")},
						metrics: []model.Metric{
							model.NewCounterMetric("id1", 123),
							model.NewGaugeMetric("id2", 1.23),
						},
					},
					calls: calls{
						DecodeBatch: 1,
						StoreBatch:  0,
						Close:       1,
					},
					errs: errs{
						DecodeBatch: errors.New("decode failed"),
						StoreBatch:  nil,
						Close:       errors.New("oops"),
					},
					want: want{
						dirtyWrites: 0,
					},
					assertion: assert.Error,
				},
			},
		},
	}
	for _, outer := range tests {
		t.Run(outer.name, func(t *testing.T) {
			for _, tt := range outer.cases {
				t.Run(tt.name, func(t *testing.T) {
					p := &metricSnapshotter{
						MetricStorer: tt.fields.MetricStorer,
						decoder:      tt.fields.decoder,
						notifyCh:     make(chan struct{}, 1),
					}
					tt.fields.decoder.metrics = tt.args.metrics
					{
						mc := tt.fields.MetricStorer.On("StoreBatch", tt.args.metrics).Return(tt.errs.StoreBatch).Times(tt.calls.StoreBatch)
						if tt.calls.StoreBatch == 0 {
							mc.Maybe()
						}
					}
					{
						mc := tt.fields.decoder.On("DecodeBatch", tt.args.r).Return(tt.args.metrics, tt.errs.DecodeBatch).Times(tt.calls.DecodeBatch)
						if tt.calls.DecodeBatch == 0 {
							mc.Maybe()
						}
					}
					{
						mc := tt.args.r.On("Close").Return(tt.errs.Close).Times(tt.calls.Close)
						if tt.calls.Close == 0 {
							mc.Maybe()
						}
					}

					err := p.LoadClose(tt.args.r)

					tt.assertion(t, err)
					tt.fields.MetricStorer.AssertExpectations(t)
					tt.fields.decoder.AssertExpectations(t)
					tt.args.r.AssertExpectations(t)
					assert.Equal(t, tt.want.dirtyWrites, p.dirtyWrites.Load())
				})
			}
		})
	}
}

func Test_metricSnapshotter_C(t *testing.T) {
	type fields struct {
		notifyCh chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "buffered",
			fields:  fields{notifyCh: make(chan struct{}, 1)},
			wantErr: false,
		},
		{
			name:    "non buffered",
			fields:  fields{notifyCh: make(chan struct{})},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &metricSnapshotter{
				notifyCh: tt.fields.notifyCh,
			}
			got := p.C()
			select {
			case tt.fields.notifyCh <- struct{}{}:
				assert.True(t, !tt.wantErr)
			default:
				assert.Truef(t, tt.wantErr, "failed to write to the channel")
			}
			select {
			case v := <-got:
				assert.Equal(t, struct{}{}, v)
			default:
				assert.Truef(t, tt.wantErr, "failed to read from the channel")
			}
		})
	}
}
