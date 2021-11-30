DROP INDEX IF EXISTS unique_ig_code;

ALTER TABLE posts
DROP COLUMN instagram_code;
