-- Add is_favourite column to posts table
ALTER TABLE photos.posts ADD COLUMN IF NOT EXISTS is_favourite BOOLEAN NOT NULL DEFAULT FALSE;