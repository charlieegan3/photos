-- https://gist.github.com/abn/779166b0c766ce67351c588489831852
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE OR REPLACE FUNCTION public.slugify(
  v TEXT
) RETURNS TEXT
  LANGUAGE plpgsql
  STRICT IMMUTABLE AS
$function$
BEGIN
  -- 1. trim trailing and leading whitespaces from text
  -- 2. remove accents (diacritic signs) from a given text
  -- 3. lowercase unaccented text
  -- 4. replace non-alphanumeric (excluding hyphen, underscore) with a hyphen
  -- 5. trim leading and trailing hyphens
  RETURN trim(BOTH '-' FROM regexp_replace(lower(unaccent(trim(v))), '[^a-z0-9\-_]+', '-', 'gi'));
END;
$function$;

CREATE OR REPLACE FUNCTION public.set_slug_from_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.slug := slugify(NEW.name);
  RETURN NEW;
END
$$;

CREATE TRIGGER set_slug_insert
BEFORE INSERT ON devices
FOR EACH ROW
EXECUTE PROCEDURE set_slug_from_name();

CREATE TRIGGER set_slug_update
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE PROCEDURE set_slug_from_name();

ALTER TABLE devices ADD COLUMN slug text;
ALTER TABLE devices DROP COLUMN icon_url;
