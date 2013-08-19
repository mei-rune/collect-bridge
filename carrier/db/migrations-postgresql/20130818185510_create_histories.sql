-- +goose Up
CREATE TABLE IF NOT EXISTS tpt_histories (
  id                BIGSERIAL  PRIMARY KEY,
  action_id         bigint  NOT NULL,
  managed_type      varchar(200)  NOT NULL,
  managed_id        bigint  NOT NULL,
  current_value     NUMERIC(20, 4)  NOT NULL,
  sampled_at        timestamp with time zone
);
CREATE INDEX tpt_histories_mo_id_idx ON tpt_histories USING btree (managed_type, managed_id);
CREATE INDEX tpt_histories_action_id_idx ON tpt_histories USING btree (action_id);


-- +goose Down
DROP INDEX IF EXISTS tpt_histories_mo_id_idx;
DROP INDEX IF EXISTS tpt_histories_action_id_idx;
DROP TABLE IF EXISTS tpt_histories CASCADE;