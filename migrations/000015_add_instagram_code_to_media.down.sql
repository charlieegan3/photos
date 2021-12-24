DROP INDEX IF EXISTS media_unique_ig_code;

ALTER TABLE medias
DROP COLUMN instagram_code;
