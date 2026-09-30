-- +goose Up
CREATE TABLE IF NOT EXISTS events
(
    id          String,
    type        LowCardinality(String),
    status      LowCardinality(String),
    occurred_at DateTime64(3),
    received_at DateTime64(3)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (type, status, occurred_at)
SETTINGS index_granularity = 8192;

-- +goose Down
DROP TABLE IF EXISTS events;