ALTER TABLE taggings DROP CONSTRAINT fk_tag_id;
ALTER TABLE taggings ADD CONSTRAINT fk_tag_id FOREIGN KEY(tag_id) REFERENCES tags(id);