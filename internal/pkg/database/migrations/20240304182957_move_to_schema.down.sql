
ALTER TABLE IF EXISTS "photos".devices
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".lenses
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".locations
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".medias
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".posts
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".taggings
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".tags
    SET SCHEMA public;

ALTER TABLE IF EXISTS "photos".trips
    SET SCHEMA public;

DROP SCHEMA IF EXISTS photos;
