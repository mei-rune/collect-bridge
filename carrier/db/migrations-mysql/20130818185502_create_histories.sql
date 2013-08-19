-- +goose Up
CREATE TABLE IF NOT EXISTS tpt_histories (
  id                SERIAL PRIMARY KEY,
  action_id         bigint  NOT NULL,
  managed_type      varchar(200)  NOT NULL,
  managed_id        bigint  NOT NULL,
  current_value     NUMERIC(20, 4)  NOT NULL,
  sampled_at        timestamp NOT NULL
);
CREATE INDEX tpt_histories_mo_id_idx ON tpt_histories (managed_type, managed_id);
CREATE INDEX tpt_histories_action_id_idx ON tpt_histories (action_id);

-- +goose Down
DROP TABLE IF EXISTS tpt_histories CASCADE;