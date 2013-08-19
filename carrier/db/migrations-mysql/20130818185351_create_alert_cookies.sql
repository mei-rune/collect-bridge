-- +goose Up
CREATE TABLE IF NOT EXISTS tpt_alert_cookies (
  id                SERIAL PRIMARY KEY,
  action_id         bigint  NOT NULL,
  managed_type      varchar(200)  NOT NULL,
  managed_id        bigint  NOT NULL,
  status            int  NOT NULL,
  previous_status   int  NOT NULL,
  event_id          varchar(200)  NOT NULL,
  sequence_id       int  NOT NULL,
  content           varchar(250)  NOT NULL,
  current_value     varchar(2000)  NOT NULL,
  triggered_at      timestamp NOT NULL,
  CONSTRAINT tpt_alert_cookies_unique_action_id UNIQUE (action_id)
);
CREATE INDEX tpt_alert_cookies_action_id_idx ON tpt_alert_cookies (action_id);

-- +goose Down
DROP TABLE IF EXISTS tpt_alert_cookies CASCADE;