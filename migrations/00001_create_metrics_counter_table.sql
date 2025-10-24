--=================================================================================
-- UP
--=================================================================================
-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics_counter (
    metric_id varchar(255) PRIMARY KEY,
    value integer NOT NULL
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
DROP TABLE IF EXISTS metrics_counter;

-- +goose StatementEnd
