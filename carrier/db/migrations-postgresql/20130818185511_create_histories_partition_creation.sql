-- +goose Up
CREATE OR REPLACE FUNCTION tpt_histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
  create_query text;                   -- declare create_query variable
  drop_sampled_at_index_query text;    -- declare drop_sampled_at_index_query variable
  create_sampled_at_index_query text;  -- declare create_sampled_at_index_query variable
  drop_mo_index_query text;            -- declare drop_mo_index_query variable
  create_mo_index_query text;          -- declare create_mo_index_query variable
  drop_action_index_query text;        -- declare drop_action_index_query variable
  create_action_index_query text;      -- declare create_action_index_query variable
BEGIN
  FOR create_query, drop_sampled_at_index_query, create_sampled_at_index_query,
  drop_mo_index_query, create_mo_index_query,
  drop_action_index_query, create_action_index_query IN SELECT
      'CREATE TABLE IF NOT EXISTS tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' ( CHECK( sampled_at >= timestamp '''
      || TO_CHAR( d, 'YYYY-MM-DD 00:00:00' )
      || ''' AND sampled_at < timestamp '''
      || TO_CHAR( d + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
      || ''' ) ) inherits ( tpt_histories );',
      'DROP INDEX  IF EXISTS tpt_histories_'
      ||TO_CHAR( d, 'YYYY_MM' )
      ||'_sampled_at_idx;',
      'CREATE INDEX tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || '_sampled_at_idx ON tpt_histories_' 
      || TO_CHAR( d, 'YYYY_MM' )
      || ' ( sampled_at );',
       'DROP INDEX IF EXISTS tpt_histories_'
       || TO_CHAR( d, 'YYYY_MM' )
       || '_mo_id_idx;',
      'CREATE INDEX tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || '_mo_id_idx ON tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' USING btree (managed_type COLLATE pg_catalog."default", managed_id);',
      'DROP INDEX IF EXISTS tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || '_action_id_idx;',
      'CREATE INDEX tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || '_action_id_idx ON tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' USING btree (action_id);'
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    EXECUTE create_query;                   -- excute create table
    EXECUTE drop_sampled_at_index_query;    -- drop index on triggered_at
    EXECUTE create_sampled_at_index_query;  -- create index on triggered_at
    EXECUTE drop_mo_index_query;            -- drop index on mo
    EXECUTE create_mo_index_query;          -- create index on mo
    EXECUTE drop_action_index_query;        -- drop index on action
    EXECUTE create_action_index_query;      -- create index on action
  END LOOP;  -- LOOP END
END;         -- FUNCTION END
$$
language plpgsql;

CREATE OR REPLACE FUNCTION tpt_histories_partition_deletion( DATE, DATE )
returns void AS $$
DECLARE
  create_query text; -- declare create table varible
BEGIN
  FOR create_query IN SELECT
      'DROP TABLE IF EXISTS tpt_histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' CASCADE;'
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    EXECUTE create_query; -- execute sql
  END LOOP; -- LOOP END
END; -- FUNC END
$$
language plpgsql;


-- +goose Down
DROP FUNCTION tpt_histories_partition_deletion( DATE, DATE);
DROP FUNCTION tpt_histories_partition_creation( DATE, DATE);