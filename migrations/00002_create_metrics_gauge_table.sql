--=================================================================================
-- UP
--=================================================================================
-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics_gauge (
    metric_id varchar(255) PRIMARY KEY,
    value double precision NOT NULL
);

-- +goose StatementEnd
--
--
--
--=================================================================================
-- DOWN
--=================================================================================
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics_gauge;

-- +goose StatementEnd
