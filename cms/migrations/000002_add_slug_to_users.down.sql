-- https://gist.github.com/abn/779166b0c766ce67351c588489831852
ALTER TABLE devices DROP COLUMN slug;
ALTER TABLE devices ADD COLUMN icon_url text CONSTRAINT icon_url_present CHECK ((icon_url != '') IS TRUE);

DROP TRIGGER IF EXISTS set_slug_insert ON devices;
DROP TRIGGER IF EXISTS set_slug_update ON devices;

DROP FUNCTION IF EXISTS public.set_slug_from_name();
DROP FUNCTION IF EXISTS public.slugify();

DROP EXTENSION IF EXISTS unaccent;
