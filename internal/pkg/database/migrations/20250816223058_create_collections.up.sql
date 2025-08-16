-- Create collections table
CREATE TABLE photos.collections (
  id SERIAL NOT NULL PRIMARY KEY,
  title text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add timestamp trigger for updated_at
CREATE TRIGGER set_timestamp_update
    BEFORE UPDATE ON photos.collections
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();