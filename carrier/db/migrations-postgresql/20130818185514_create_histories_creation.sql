-- +goose Up
SELECT tpt_histories_partition_creation( '2010-01-01', '2028-01-01' );


-- +goose Down
SELECT tpt_histories_partition_deletion( '2010-01-01', '2028-01-01' );