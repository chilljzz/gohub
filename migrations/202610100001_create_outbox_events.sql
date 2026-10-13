
-- +goose NO TRANSACTION

-- +goose Up

CREATE TABLE outbox_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    event_id CHAR(36) NOT NULL,

    aggregate_type VARCHAR(32) NOT NULL,
    aggregate_id BIGINT UNSIGNED NOT NULL,

    event_type VARCHAR(64) NOT NULL,
    payload JSON NOT NULL,

    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    retry_count INT UNSIGNED NOT NULL DEFAULT 0,

    next_retry_at DATETIME(3)
        NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    locked_until DATETIME(3) NULL,
    last_error VARCHAR(500) NULL,

    created_at DATETIME(3)
        NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    updated_at DATETIME(3)
        NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
        ON UPDATE CURRENT_TIMESTAMP(3),

    published_at DATETIME(3) NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uk_outbox_event_id (event_id),

    KEY idx_outbox_poll (
        status, next_retry_at, id
    ),

    KEY idx_outbox_recovery (
        status, locked_until, id
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down

DROP TABLE outbox_events;
