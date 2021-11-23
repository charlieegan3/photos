-- medias stores information about media files uploaded to the system
CREATE TABLE medias (
  id SERIAL NOT NULL PRIMARY KEY,

  make text NOT NULL DEFAULT '',
  model text NOT NULL DEFAULT '',

  taken_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  f_number float NOT NULL DEFAULT 0,
  shutter_speed float NOT NULL DEFAULT 0,
  iso_speed int NOT NULL DEFAULT 0,

  latitude float NOT NULL DEFAULT 0,
  longitude float NOT NULL DEFAULT 0,
  altitude float NOT NULL DEFAULT 0,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- updated_at is set by trigger_set_timestamp
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON medias
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON medias
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
