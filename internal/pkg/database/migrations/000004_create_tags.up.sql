-- tags stores references to a single tag used on a post
CREATE TABLE tags (
  id SERIAL NOT NULL PRIMARY KEY,

  name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),

  hidden BOOLEAN NOT NULL DEFAULT FALSE,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- updated_at is set by trigger_set_timestamp
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE (name)
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON tags
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON tags
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE OR REPLACE FUNCTION public.slugify_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.name := slugify(NEW.name);
  RETURN NEW;
END
$$;

CREATE TRIGGER slugify_name_update
BEFORE UPDATE ON tags
FOR EACH ROW
EXECUTE PROCEDURE slugify_name();

CREATE TRIGGER slugify_name_insert
BEFORE INSERT ON tags
FOR EACH ROW
EXECUTE PROCEDURE slugify_name();
