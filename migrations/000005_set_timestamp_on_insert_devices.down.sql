CREATE TRIGGER set_timestamp
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

DROP TRIGGER IF EXISTS set_timestamp_update ON devices;
DROP TRIGGER IF EXISTS set_timestamp_insert ON devices;
