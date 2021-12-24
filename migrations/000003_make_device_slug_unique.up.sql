ALTER TABLE devices
ADD CONSTRAINT slug_unique UNIQUE (slug);
