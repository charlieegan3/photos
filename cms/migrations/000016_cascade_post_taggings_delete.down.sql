ALTER TABLE taggings DROP CONSTRAINT fk_post_id;
ALTER TABLE taggings ADD CONSTRAINT fk_post_id FOREIGN KEY(post_id) REFERENCES posts(id);
