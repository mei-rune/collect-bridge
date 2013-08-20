-- +goose Up
CREATE OR REPLACE FUNCTION tpt_alert_histories_partition_function()
returns TRIGGER AS $$
begin
execute 'INSERT INTO tpt_alert_histories_'
|| to_char( NEW.triggered_at, 'YYYY_MM' )
|| '(action_id, managed_type, managed_id, status, previous_status, event_id, sequence_id, content, current_value, triggered_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)'
USING NEW.action_id, NEW.managed_type, NEW.managed_id, NEW.status, NEW.previous_status, NEW.event_id, NEW.sequence_id, NEW.content, NEW.current_value, NEW.triggered_at; --
RETURN NULL; -- return
end; -- end func
$$
LANGUAGE plpgsql;

CREATE TRIGGER tpt_alert_histories_partition_trigger
before INSERT
ON tpt_alert_histories
FOR each row
execute procedure tpt_alert_histories_partition_function() ;


-- +goose Down
DROP TRIGGER IF EXISTS tpt_alert_histories_partition_trigger ON tpt_alert_histories;
DROP FUNCTION tpt_alert_histories_partition_function();