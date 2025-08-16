-- Create post_collections junction table
CREATE TABLE photos.post_collections (
  id SERIAL NOT NULL PRIMARY KEY,
  post_id INT NOT NULL,
  collection_id INT NOT NULL,
  
  CONSTRAINT fk_post_id FOREIGN KEY(post_id) REFERENCES photos.posts(id) ON DELETE CASCADE,
  CONSTRAINT fk_collection_id FOREIGN KEY(collection_id) REFERENCES photos.collections(id) ON DELETE CASCADE,
  
  UNIQUE (post_id, collection_id)
);