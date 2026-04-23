package sqlstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bq2cd/yp-go-metrics/internal/model"
)

func Test_sqlItemCounter_GetID(t *testing.T) {
	type fields struct {
		ID    string
		Value int64
	}
	type want struct {
		got string
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemCounter{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.GetID()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemCounter_GetValue(t *testing.T) {
	type fields struct {
		ID    string
		Value int64
	}
	type want struct {
		got any
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemCounter{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.GetValue()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemCounter_MetricType(t *testing.T) {
	type fields struct {
		ID    string
		Value int64
	}
	type want struct {
		got model.MetricType
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemCounter{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.MetricType()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemCounter_ToMetric(t *testing.T) {
	type fields struct {
		ID    string
		Value int64
	}
	type want struct {
		got model.Metric
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemCounter{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.ToMetric()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemCounter_FromMetric(t *testing.T) {
	type fields struct {
		ID    string
		Value int64
	}
	type args struct {
		m model.Metric
	}
	type want struct {
		got sqlItem
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
			it := sqlItemCounter{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.FromMetric(tt.args.m)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemGauge_GetID(t *testing.T) {
	type fields struct {
		ID    string
		Value float64
	}
	type want struct {
		got string
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemGauge{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.GetID()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemGauge_GetValue(t *testing.T) {
	type fields struct {
		ID    string
		Value float64
	}
	type want struct {
		got any
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemGauge{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.GetValue()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemGauge_MetricType(t *testing.T) {
	type fields struct {
		ID    string
		Value float64
	}
	type want struct {
		got model.MetricType
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemGauge{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.MetricType()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemGauge_ToMetric(t *testing.T) {
	type fields struct {
		ID    string
		Value float64
	}
	type want struct {
		got model.Metric
	}
	type testcase struct {
		fields fields
		want   want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			it := sqlItemGauge{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.ToMetric()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func Test_sqlItemGauge_FromMetric(t *testing.T) {
	type fields struct {
		ID    string
		Value float64
	}
	type args struct {
		m model.Metric
	}
	type want struct {
		got sqlItem
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
			it := sqlItemGauge{
				ID:    tt.fields.ID,
				Value: tt.fields.Value,
			}
			got := it.FromMetric(tt.args.m)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
