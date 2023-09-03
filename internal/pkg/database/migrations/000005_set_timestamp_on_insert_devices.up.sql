DROP TRIGGER IF EXISTS set_timestamp ON devices;

CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON devices
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
