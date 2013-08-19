-- +goose Up
IF object_id('dbo.tpt_histories', 'U') IS NULL
BEGIN
  CREATE TABLE tpt_histories (
    id                BIGINT IDENTITY(1,1)   PRIMARY KEY,
    action_id         bigint  NOT NULL,
    managed_type      varchar(200)  NOT NULL,
    managed_id        bigint  NOT NULL,
    current_value     NUMERIC(20, 4)  NOT NULL,
    sampled_at        datetime NOT NULL
  );                                                                            -- create table
  CREATE INDEX tpt_histories_mo_id_idx ON tpt_histories (managed_type, managed_id); --
  CREATE INDEX tpt_histories_action_id_idx ON tpt_histories (action_id);            --
END;

-- +goose Down
IF object_id('dbo.tpt_histories', 'U') IS NOT NULL
BEGIN
  DROP TABLE tpt_histories; --
END;