-- +goose Up
IF object_id('dbo.tpt_alert_cookies', 'U') IS NULL
BEGIN
  CREATE TABLE tpt_alert_cookies (
    id                BIGINT IDENTITY(1,1)   PRIMARY KEY,
    action_id         bigint  NOT NULL,
    managed_type      varchar(200)  NOT NULL,
    managed_id        bigint  NOT NULL,
    status            int  NOT NULL,
    previous_status   int  NOT NULL,
    event_id          varchar(200)  NOT NULL,
    sequence_id       int  NOT NULL,
    level             int  NOT NULL,
    content           varchar(250)  NOT NULL,
    current_value     varchar(2000)  NOT NULL,
    triggered_at      datetime NOT NULL,
    CONSTRAINT tpt_alert_cookies_unique_action_id UNIQUE (action_id)
  );                                                                            -- create table
  CREATE INDEX tpt_alert_cookies_action_id_idx ON tpt_alert_cookies (action_id); -- create index
END;

-- +goose Down
IF object_id('dbo.tpt_alert_cookies', 'U') IS NOT NULL
BEGIN
  DROP TABLE tpt_alert_cookies;                -- drop table
END;