-- locations store references to places which media can be associated with
CREATE TABLE locations (
  id SERIAL NOT NULL PRIMARY KEY,

  name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),
  slug text NOT NULL DEFAULT '',

  latitude float NOT NULL DEFAULT 0,
  longitude float NOT NULL DEFAULT 0,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- updated_at is set by trigger_set_timestamp
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE (name)
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON locations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON locations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_slug_insert
BEFORE INSERT ON locations
FOR EACH ROW
EXECUTE PROCEDURE set_slug_from_name();

CREATE TRIGGER set_slug_update
BEFORE UPDATE ON locations
FOR EACH ROW
EXECUTE PROCEDURE set_slug_from_name();
