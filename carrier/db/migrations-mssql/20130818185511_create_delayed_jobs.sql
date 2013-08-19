-- +goose Up
IF object_id('dbo.tpt_delayed_jobs', 'U') IS NULL
BEGIN
    CREATE TABLE tpt_delayed_jobs (
          id                BIGINT IDENTITY(1,1)  PRIMARY KEY,
          priority          int DEFAULT 0,
          attempts          int DEFAULT 0,
          queue             varchar(200),
          handler           text  NOT NULL,
          handler_id        varchar(200),
          last_error        varchar(2000),
          run_at            DATETIME2,
          locked_at         DATETIME2,
          failed_at         DATETIME2,
          locked_by         varchar(200),
          created_at        DATETIME2 NOT NULL,
          updated_at        DATETIME2 NOT NULL
    ); --
END;

-- +goose Down
IF object_id('dbo.tpt_delayed_jobs', 'U') IS NOT NULL
BEGIN
  DROP TABLE tpt_delayed_jobs; --
END;