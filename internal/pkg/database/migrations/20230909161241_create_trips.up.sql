CREATE TABLE trips (
  id SERIAL NOT NULL PRIMARY KEY,

  title text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',

  start_date TIMESTAMPTZ NOT NULL,
  end_date TIMESTAMPTZ NOT NULL,


  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- updated_at is set by trigger_set_timestamp
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
    BEFORE UPDATE ON trips
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
    BEFORE INSERT ON trips
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
