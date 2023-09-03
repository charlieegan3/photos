-- devices stores records for devices used to take photos such as cameras and
-- phones
CREATE TABLE devices (
  id SERIAL NOT NULL PRIMARY KEY,

  name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),
  icon_url text CONSTRAINT icon_url_present CHECK ((icon_url != '') IS TRUE),

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- updated_at is set by trigger_set_timestamp
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE (name)
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
