-- +goose Up
CREATE OR REPLACE FUNCTION tpt_histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
  t DATE;     -- the time
  tname text; -- the table name
BEGIN
  FOR t, tname IN SELECT d, 'tpt_histories_' || TO_CHAR( d, 'YYYY_MM' )
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    IF NOT EXISTS (SELECT * FROM pg_catalog.pg_tables WHERE tablename = tname) THEN
      EXECUTE  'CREATE TABLE '
        || tname
        || ' ( CHECK( sampled_at >= timestamp '''
        || TO_CHAR( t, 'YYYY-MM-DD 00:00:00' )
        || ''' AND sampled_at < timestamp '''
        || TO_CHAR( t + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
        || ''' ) ) inherits ( tpt_histories );'
        || 'CREATE INDEX '
        || tname
        || '_sampled_at_idx ON ' 
        || tname
        || ' ( sampled_at );'
        || 'CREATE INDEX '
        || tname
        || '_mo_id_idx ON '
        || tname
        || ' USING btree (managed_type COLLATE pg_catalog."default", managed_id);'
        || 'CREATE INDEX '
        || tname
        || '_action_id_idx ON '
        || tname
        || ' USING btree (action_id);';  -- create table
    END IF;  -- IF END
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