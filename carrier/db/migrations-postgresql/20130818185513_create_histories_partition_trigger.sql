-- +goose Up
CREATE OR REPLACE FUNCTION tpt_histories_partition_function()
returns TRIGGER AS $$
begin
execute 'INSERT INTO tpt_histories_'
|| to_char( NEW.sampled_at, 'YYYY_MM' )
|| '(action_id, managed_type, managed_id, current_value, sampled_at) VALUES ($1, $2, $3, $4, $5)' USING NEW.action_id, NEW.managed_type, NEW.managed_id, NEW.current_value, NEW.sampled_at ; --
RETURN NULL; -- return
end; -- func end
$$
LANGUAGE plpgsql;

CREATE TRIGGER tpt_histories_partition_trigger
before INSERT
ON tpt_histories
FOR each row
execute procedure tpt_histories_partition_function() ;


-- +goose Down
DROP TRIGGER IF EXISTS tpt_histories_partition_trigger ON tpt_histories;
drop function tpt_histories_partition_function();