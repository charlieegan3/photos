-- Remove is_favourite column from posts table
ALTER TABLE photos.posts DROP COLUMN IF EXISTS is_favourite;