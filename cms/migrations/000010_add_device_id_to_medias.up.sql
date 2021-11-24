ALTER TABLE medias
ADD COLUMN device_id int NOT NULL,
ADD CONSTRAINT fk_device_id
FOREIGN KEY (device_id)
REFERENCES devices (id);
