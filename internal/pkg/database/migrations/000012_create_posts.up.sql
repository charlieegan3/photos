-- posts stores records of posts to be viewed by visitors to the site
CREATE TABLE posts (
  id SERIAL NOT NULL PRIMARY KEY,

  media_id INT NOT NULL,
  location_id INT NOT NULL,

  description text NOT NULL DEFAULT '',

  publish_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  is_draft BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT fk_media_id
     FOREIGN KEY(media_id)
     REFERENCES medias(id),

  CONSTRAINT fk_location_id
     FOREIGN KEY(location_id)
     REFERENCES locations(id),

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp_update
BEFORE UPDATE ON posts
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_insert
BEFORE INSERT ON posts
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
