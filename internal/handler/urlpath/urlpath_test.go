package urlpath

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewOperationFromURLPath(t *testing.T) {
	type args struct {
		urlPath string
	}
	tests := []struct {
		name string
		args args
		want MetricOperation
	}{
		// no operation
		{
			name: "empty path",
			args: args{urlPath: ""},
			want: MetricOperation{},
		},
		{
			name: "no operation",
			args: args{urlPath: "/"},
			want: MetricOperation{},
		},
		{
			name: "no operation 2",
			args: args{urlPath: "///"},
			want: MetricOperation{},
		},
		{
			name: "no operation 3",
			args: args{urlPath: " / / / "},
			want: MetricOperation{},
		},
		// some operation
		{
			name: "some operation",
			args: args{urlPath: "op1"},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 2",
			args: args{urlPath: "/op1"},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 3",
			args: args{urlPath: "op1/"},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 4",
			args: args{urlPath: "/op1/"},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 5",
			args: args{urlPath: " / op1 / "},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 6",
			args: args{urlPath: " op1 / "},
			want: MetricOperation{Type: OperationType("op1")},
		},
		{
			name: "some operation 7",
			args: args{urlPath: "  op1  "},
			want: MetricOperation{Type: OperationType("op1")},
		},
		// operation + type
		{
			name: "operation + type",
			args: args{urlPath: "/op1/type1"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1")},
		},
		{
			name: "operation + type 2",
			args: args{urlPath: "/op1/type1/"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1")},
		},
		{
			name: "operation + type 3",
			args: args{urlPath: "op1/type1/"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1")},
		},
		{
			name: "operation + type 4",
			args: args{urlPath: "op1/type1"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1")},
		},
		{
			name: "operation + type 5",
			args: args{urlPath: " op1  / type1  "},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1")},
		},
		// operation + type + id
		{
			name: "operation + type + id",
			args: args{urlPath: "/op1/type1/id1"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1"},
		},
		{
			name: "operation + type + id 2",
			args: args{urlPath: "/op1/type1/id1/"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1"},
		},
		{
			name: "operation + type + id 3",
			args: args{urlPath: "op1/type1/id1/"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1"},
		},
		{
			name: "operation + type + id 4",
			args: args{urlPath: "op1/type1/id1"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1"},
		},
		{
			name: "operation + type + id 5",
			args: args{urlPath: " // op1 // type1 // id1 // "},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1"},
		},
		// operation + type + id + value
		{
			name: "operation + type + id + value",
			args: args{urlPath: "/op1/type1/id1/-none.23"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "-none.23"},
		},
		{
			name: "operation + type + id + value 2",
			args: args{urlPath: "/op1/type1/id1/-none.23//"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "-none.23"},
		},
		{
			name: "operation + type + id + value 3",
			args: args{urlPath: " op1/type1/id1/-none.23//"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "-none.23"},
		},
		{
			name: "operation + type + id + value 4",
			args: args{urlPath: " op1/type1/id1/ -none.23 "},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "-none.23"},
		},
		{
			name: "operation + type + id + value 5",
			args: args{urlPath: "  //// op1    // type1   /////   id1 //   /   /  -none.23  //   /"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "/   /  -none.23  /"},
		},
		{
			name: "operation + type + id + value 6",
			args: args{urlPath: "   op1  /   type1 /  id1 /  -none.23   "},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "-none.23"},
		},
		{
			name: "extra path components",
			args: args{urlPath: "/op1/type1/id1/123/bla"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "123/bla"},
		},
		{
			name: "extra path components 2",
			args: args{urlPath: "/op1/type1/id1/123/bla//none"},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "123/bla/none"},
		},
		{
			name: "extra path components 3",
			args: args{urlPath: "/op1/type1/id1/123/bla/ /none/// "},
			want: MetricOperation{Type: OperationType("op1"), MetricType: model.MetricType("type1"), MetricID: "id1", MetricValue: "123/bla/ /none"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewOperationFromURLPath(tt.args.urlPath))
		})
	}
}

func TestMetricOperation_ToMetric(t *testing.T) {
	type fields struct {
		Type        OperationType
		MetricType  model.MetricType
		MetricID    string
		MetricValue string
	}
	tests := []struct {
		name      string
		fields    fields
		want      model.Metric
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricType)
			},
		},
		{
			name:   "valid operation",
			fields: fields{Type: OperationTypeUpdate},
			want:   model.Metric{},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricType)
			},
		},
		{
			name:   "valid metric type",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter},
			want:   model.Metric{Type: model.MetricTypeCounter},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricID)
			},
		},
		{
			name:   "some metric id",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1"},
			want:   model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricValue)
			},
		},
		{
			name:   "invalid counter value",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1", MetricValue: "-2+none"},
			want:   model.Metric{Type: model.MetricTypeCounter, ID: "id1"},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrInvalidMetricValue)
			},
		},
		{
			name:   "valid counter value",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1", MetricValue: "-123"},
			want:   model.NewCounterMetric("id1", -123),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "invalid gauge value",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeGauge, MetricID: "id1", MetricValue: "-2.34.none"},
			want:   model.Metric{Type: model.MetricTypeGauge, ID: "id1"},
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrInvalidMetricValue)
			},
		},
		{
			name:   "valid gauge value",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeGauge, MetricID: "id1", MetricValue: "-1.23"},
			want:   model.NewGaugeMetric("id1", -1.23),
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := &MetricOperation{
				Type:        tt.fields.Type,
				MetricType:  tt.fields.MetricType,
				MetricID:    tt.fields.MetricID,
				MetricValue: tt.fields.MetricValue,
			}
			got, err := mo.ToMetric()
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewOperationFromMetric(t *testing.T) {
	type args struct {
		operation OperationType
		metric    model.Metric
	}
	tests := []struct {
		name string
		args args
		want MetricOperation
	}{
		{
			name: "empty operation + empty metric",
			args: args{operation: OperationType(""), metric: model.Metric{}},
			want: MetricOperation{},
		},
		{
			name: "empty metric",
			args: args{operation: OperationTypeUpdate, metric: model.Metric{}},
			want: MetricOperation{Type: OperationTypeUpdate},
		},
		{
			name: "metric without id",
			args: args{operation: OperationTypeUpdate, metric: model.Metric{Type: model.MetricTypeCounter}},
			want: MetricOperation{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter},
		},
		{
			name: "metric without value",
			args: args{operation: OperationTypeUpdate, metric: model.Metric{Type: model.MetricTypeCounter, ID: "id1"}},
			want: MetricOperation{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1"},
		},
		{
			name: "valid counter",
			args: args{operation: OperationTypeUpdate, metric: model.NewCounterMetric("id1", 123)},
			want: MetricOperation{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1", MetricValue: "123"},
		},
		{
			name: "valid gauge",
			args: args{operation: OperationTypeUpdate, metric: model.NewGaugeMetric("id1", -1.23)},
			want: MetricOperation{Type: OperationTypeUpdate, MetricType: model.MetricTypeGauge, MetricID: "id1", MetricValue: "-1.23"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewOperationFromMetric(tt.args.operation, tt.args.metric))
		})
	}
}

func TestMetricOperation_ToURLPath(t *testing.T) {
	type fields struct {
		Type        OperationType
		MetricType  model.MetricType
		MetricID    string
		MetricValue string
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		// error
		{
			name:   "empty operation",
			fields: fields{},
			want:   "",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingOperation)
			},
		},
		{
			name:   "some operation",
			fields: fields{Type: OperationTypeUpdate},
			want:   "",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.Error(t, err, ErrMissingMetricType)
			},
		},
		{
			name:   "operation + metric type",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter},
			want:   "",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricID)
			},
		},
		{
			name:   "update metric without value",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1"},
			want:   "",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrMissingMetricValue)
			},
		},
		{
			name:   "invalid operation",
			fields: fields{Type: OperationType("invalid"), MetricType: model.MetricTypeCounter, MetricID: "id1", MetricValue: "456"},
			want:   "",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.ErrorIs(t, err, ErrInvalidOperation)
			},
		},
		// ok
		{
			name:   "read metric with some id",
			fields: fields{Type: OperationTypeValue, MetricType: model.MetricTypeCounter, MetricID: "id1"},
			want:   "/value/counter/id1",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "operation + valid counter",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeCounter, MetricID: "id1", MetricValue: "-15"},
			want:   "/update/counter/id1/-15",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:   "operation + valid gauge",
			fields: fields{Type: OperationTypeUpdate, MetricType: model.MetricTypeGauge, MetricID: "id1", MetricValue: "-1.56"},
			want:   "/update/gauge/id1/-1.56",
			assertion: func(t assert.TestingT, err error, v ...any) bool {
				return assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mo := &MetricOperation{
				Type:        tt.fields.Type,
				MetricType:  tt.fields.MetricType,
				MetricID:    tt.fields.MetricID,
				MetricValue: tt.fields.MetricValue,
			}
			got, err := mo.ToURLPath()
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
