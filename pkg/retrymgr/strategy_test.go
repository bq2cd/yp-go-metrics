package retrymgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStrategy1s3s5s(t *testing.T) {
	got := NewStrategy1s3s5s()
	s, ok := got.(*strategy1s3s5s)
	require.Truef(t, ok, "must be strategy1s3s5s struct")
	assert.Equal(t, [3]int{1, 3, 5}, s.steps)
	assert.Zero(t, s.current)
}

func Test_strategy1s3s5s_NextDelay(t *testing.T) {
	wantSteps := [4]time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
		0,
	}
	wantActive := [4]bool{
		true,
		true,
		true,
		false,
	}

	s := NewStrategy1s3s5s()

	for i := range wantSteps {
		delay, active := s.NextDelay()
		assert.Equal(t, wantSteps[i], delay)
		assert.Equal(t, wantActive[i], active)
	}
}

func Test_strategy1s3s5s_Name(t *testing.T) {
	assert.Equal(t, "1s3s5s", NewStrategy1s3s5s().Name())
}
