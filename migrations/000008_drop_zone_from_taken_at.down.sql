ALTER TABLE medias
ALTER COLUMN taken_at TYPE TIMESTAMPTZ,
ALTER COLUMN taken_at SET NOT NULL,
ALTER COLUMN taken_at SET DEFAULT NOW();