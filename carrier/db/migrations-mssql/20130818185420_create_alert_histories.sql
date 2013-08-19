-- +goose Up
IF object_id('dbo.tpt_alert_histories', 'U') IS NULL
BEGIN
  CREATE TABLE tpt_alert_histories (
    id                BIGINT IDENTITY(1,1)   PRIMARY KEY,
    action_id         bigint  NOT NULL,
    managed_type      varchar(200)  NOT NULL,
    managed_id        bigint  NOT NULL,
    status            int  NOT NULL,
    previous_status   int  NOT NULL,
    event_id          varchar(200)  NOT NULL,
    sequence_id       int  NOT NULL,
    content           varchar(250)  NOT NULL,
    current_value     varchar(2000) NOT NULL,
    triggered_at      datetime  NOT NULL
  );                                                                            -- create table
  CREATE INDEX tpt_alert_histories_mo_id_idx ON tpt_alert_histories (managed_type, managed_id); --
  CREATE INDEX tpt_alert_histories_action_id_idx ON tpt_alert_histories (action_id);   --
END;

-- +goose Down
IF object_id('dbo.tpt_alert_histories', 'U') IS NOT NULL
BEGIN
  DROP TABLE tpt_alert_histories;   --
END