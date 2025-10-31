package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/bq2cd/yp-go-metrics/internal/model"
	"github.com/huandu/go-sqlbuilder"
)

var (
	sqlHandlersByMetricType map[model.MetricType]sqlHandler
)

func init() {
	sqlHandlersByMetricType = map[model.MetricType]sqlHandler{
		model.MetricTypeCounter: newSQLHandler[sqlItemCounter](),
		model.MetricTypeGauge:   newSQLHandler[sqlItemGauge](),
	}
}

type sqlHandler interface {
	Select(ctx context.Context, selector sqlSelector, itemIds ...string) ([]sqlItem, error)
	Insert(ctx context.Context, execer sqlExecer, items ...sqlItem) (sql.Result, error)
	ConvertMetrics(metrics model.MetricSet) []sqlItem
}

type sqlHandlerImpl[T sqlItem] struct {
	tableName string
}

func newSQLHandler[T sqlItem]() sqlHandlerImpl[T] {
	return sqlHandlerImpl[T]{
		tableName: fmt.Sprintf("%s%s", tableNamePrefix, (*new(T)).MetricType()),
	}
}

func (h sqlHandlerImpl[T]) Select(ctx context.Context, selector sqlSelector, itemIds ...string) ([]sqlItem, error) {
	sb := sqlbuilder.NewSelectBuilder().
		Select(`metric_id`, `value`).
		From(h.tableName)

	if len(itemIds) > 0 {
		sb.Where(sb.In(`metric_id`, sqlAnySlice(itemIds)...))
	}

	query, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	dest := make([]T, 0)
	err := selector.SelectContext(ctx, &dest, query, args...)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot scan metrics: %w", err)
	}

	result := make([]sqlItem, len(dest))
	for i, v := range dest {
		result[i] = v
	}

	return result, nil
}

func (h sqlHandlerImpl[T]) Insert(ctx context.Context, execer sqlExecer, items ...sqlItem) (sql.Result, error) {
	if len(items) == 0 {
		return nil, nil
	}

	sb := sqlbuilder.NewInsertBuilder().
		InsertInto(h.tableName).
		Cols(`metric_id`, `value`)

	for _, m := range items {
		sb.Values(m.GetID(), m.GetValue())
	}

	query, args := sb.
		SQL(`ON CONFLICT (metric_id) DO UPDATE SET value = EXCLUDED.value`).
		BuildWithFlavor(sqlbuilder.PostgreSQL)

	res, err := execer.ExecContext(ctx, query, args...)
	if err != nil {
		return res, fmt.Errorf("cannot insert items: %w", err)
	}

	return res, nil
}

func (h sqlHandlerImpl[T]) ConvertMetrics(metrics model.MetricSet) []sqlItem {
	items := make([]sqlItem, 0, len(metrics))
	// FIXME: temporary hack to work around `sqlmock` limitations
	// when it is expecting the arguments to be passed in the same
	// order expectations are arranged.
	for _, key := range slices.SortedStableFunc(maps.Keys(metrics), func(a, b model.MetricKey) int { return a.Compare(b) }) {
		m := metrics[key]
		if m.Empty() {
			continue
		}
		it := *new(T)
		if m.Type != it.MetricType() {
			continue
		}
		items = append(items, it.FromMetric(m))
	}
	return items
}
