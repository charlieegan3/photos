ALTER TABLE medias
ADD COLUMN instagram_code TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX medias_unique_ig_code ON medias(instagram_code) WHERE TRIM(instagram_code) <> '';
