package source

import (
	"testing"

	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/extra"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats"
	"github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSources(t *testing.T) {
	tests := []struct {
		name string
		want []Source
	}{
		{
			name: "default sources",
			want: []Source{memstats.New(), extra.New(), psutil.New()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, DefaultSources())
		})
	}
}
