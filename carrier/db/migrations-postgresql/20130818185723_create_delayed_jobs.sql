-- +goose Up
CREATE TABLE IF NOT EXISTS tpt_delayed_jobs (
    id                SERIAL PRIMARY KEY,
    priority          int DEFAULT 0,
    attempts          int DEFAULT 0,
    queue             varchar(200),
    handler           text  NOT NULL,
    handler_id        varchar(200),
    last_error        varchar(2000),
    run_at            timestamp with time zone,
    locked_at         timestamp with time zone,
    failed_at         timestamp with time zone,
    locked_by         varchar(200),
    created_at        timestamp with time zone NOT NULL,
    updated_at        timestamp with time zone NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tpt_delayed_jobs