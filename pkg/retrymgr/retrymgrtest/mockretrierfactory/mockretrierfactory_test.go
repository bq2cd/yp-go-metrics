package mockretrierfactory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-metrics/pkg/log"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr"
	"github.com/bq2cd/yp-go-metrics/pkg/retrymgr/retrymgrtest"
)

func TestNewMockRetrierFactory(t *testing.T) {
	type args struct {
		ctrl *gomock.Controller
	}
	type want struct {
		got *MockRetrierFactory
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := NewMockRetrierFactory(tt.args.ctrl)
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMockRetrierFactory_GetLogger(t *testing.T) {
	type fields struct {
		Logger   log.TestLogger
		Sleeper  *retrymgrtest.MockSleeper
		Strategy *retrymgrtest.MockStrategy
	}
	type want struct {
		got log.Logger
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
			m := &MockRetrierFactory{
				Logger:   tt.fields.Logger,
				Sleeper:  tt.fields.Sleeper,
				Strategy: tt.fields.Strategy,
			}
			got := m.GetLogger()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMockRetrierFactory_GetSleeper(t *testing.T) {
	type fields struct {
		Logger   log.TestLogger
		Sleeper  *retrymgrtest.MockSleeper
		Strategy *retrymgrtest.MockStrategy
	}
	type want struct {
		got retrymgr.Sleeper
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
			m := &MockRetrierFactory{
				Logger:   tt.fields.Logger,
				Sleeper:  tt.fields.Sleeper,
				Strategy: tt.fields.Strategy,
			}
			got := m.GetSleeper()
			assert.Equal(t, tt.want.got, got)
		})
	}
}

func TestMockRetrierFactory_GetStrategy(t *testing.T) {
	type fields struct {
		Logger   log.TestLogger
		Sleeper  *retrymgrtest.MockSleeper
		Strategy *retrymgrtest.MockStrategy
	}
	type want struct {
		got retrymgr.Strategy
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
			m := &MockRetrierFactory{
				Logger:   tt.fields.Logger,
				Sleeper:  tt.fields.Sleeper,
				Strategy: tt.fields.Strategy,
			}
			got := m.GetStrategy()
			assert.Equal(t, tt.want.got, got)
		})
	}
}
