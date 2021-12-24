ALTER TABLE posts
ADD COLUMN instagram_code TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX unique_ig_code ON posts(instagram_code) WHERE TRIM(instagram_code) <> '';
