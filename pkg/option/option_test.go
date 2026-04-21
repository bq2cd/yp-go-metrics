package option_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bq2cd/yp-go-metrics/pkg/option"
)

func TestApply(t *testing.T) {
	type config struct {
		a, b, c int
	}

	withA := func(n int) option.Option[config] {
		return func(c *config) {
			c.a = n
		}
	}
	withB := func(n int) option.Option[config] {
		return func(c *config) {
			c.b = n
		}
	}
	withC := func(n int) option.Option[config] {
		return func(c *config) {
			c.c = n
		}
	}

	tests := map[string]struct {
		initial config
		opts    []option.Option[config]
		want    config
	}{
		"empty initial, no options": {
			initial: config{},
			opts:    []option.Option[config]{},
			want:    config{},
		},
		"empty initial, some options": {
			initial: config{},
			opts:    []option.Option[config]{withA(3), withC(5)},
			want:    config{a: 3, c: 5},
		},
		"preconfigured initial, options override": {
			initial: config{a: 2, b: 5},
			opts:    []option.Option[config]{withB(-3), withC(-5)},
			want:    config{a: 2, b: -3, c: -5},
		},
	}

	for tname, tc := range tests {
		t.Run(tname, func(t *testing.T) {
			option.Apply(&tc.initial, tc.opts...)

			assert.Equal(t, tc.want, tc.initial)
		})
	}
}
