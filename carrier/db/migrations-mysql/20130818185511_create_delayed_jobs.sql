-- +goose Up
CREATE TABLE IF NOT EXISTS tpt_delayed_jobs (
    id                SERIAL PRIMARY KEY,
    priority          int DEFAULT 0,
    attempts          int DEFAULT 0,
    queue             varchar(200),
    handler           text  NOT NULL,
    handler_id        varchar(200),
    last_error        VARCHAR(200),
    run_at            DATETIME,
    locked_at         DATETIME,
    failed_at         DATETIME,
    locked_by         varchar(200),
    created_at        DATETIME NOT NULL,
    updated_at        timestamp NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tpt_delayed_jobs CASCADE;