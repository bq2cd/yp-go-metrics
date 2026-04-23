package periodictask

import (
	"context"
)

// Task defines a generic task which can be run.
type Task interface {
	Run(ctx context.Context) error
}
