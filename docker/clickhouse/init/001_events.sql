CREATE TABLE IF NOT EXISTS events
(
    site_id UUID,
    token String,
    event_type LowCardinality(String),
    url String,
    referrer String,
    user_agent String,
    ip IPv6,
    country LowCardinality(String) DEFAULT '',
    city String DEFAULT '',
    device LowCardinality(String) DEFAULT '',
    browser LowCardinality(String) DEFAULT '',
    os LowCardinality(String) DEFAULT '',
    x Nullable(UInt16),
    y Nullable(UInt16),
    meta String,
    created_at DateTime64(3)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(created_at)
ORDER BY (site_id, event_type, created_at);