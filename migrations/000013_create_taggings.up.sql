-- taggings links posts and tags
CREATE TABLE taggings (
  id SERIAL NOT NULL PRIMARY KEY,

  post_id INT NOT NULL,
  tag_id INT NOT NULL,

  CONSTRAINT fk_post_id
     FOREIGN KEY(post_id)
     REFERENCES posts(id),

  CONSTRAINT fk_tag_id
     FOREIGN KEY(tag_id)
     REFERENCES tags(id),

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE (post_id, tag_id)
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON taggings
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON taggings
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
