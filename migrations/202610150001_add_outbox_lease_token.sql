-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE outbox_events
  ADD COLUMN lease_token CHAR(32) NULL AFTER locked_until;

-- +goose Down
ALTER TABLE outbox_events
  DROP COLUMN lease_token;