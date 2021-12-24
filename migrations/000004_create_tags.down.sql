DROP TABLE IF EXISTS tags;

DROP TRIGGER IF EXISTS slugify_name_update ON tags;
DROP TRIGGER IF EXISTS slugify_name_insert ON tags;

DROP FUNCTION IF EXISTS public.slugify_name();
