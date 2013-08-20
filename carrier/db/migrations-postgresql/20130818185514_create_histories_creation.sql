-- +goose Up
SELECT tpt_histories_partition_creation( '2013-01-01', '2014-01-01' );


-- +goose Down
SELECT tpt_histories_partition_deletion( '2013-01-01', '2014-01-01' );