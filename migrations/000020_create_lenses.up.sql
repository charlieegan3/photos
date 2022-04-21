-- lenses stores records for lenses with a device to create a media
CREATE TABLE lenses (
    id SERIAL NOT NULL PRIMARY KEY,

    name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- updated_at is set by trigger_set_timestamp
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (name)
);

-- trigger_set_timestamp sets updated_at when updating a record
CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON lenses
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

ALTER TABLE medias ADD COLUMN lens_id INTEGER REFERENCES lenses (id);
ALTER TABLE medias
    ADD CONSTRAINT fk_medias_lenses FOREIGN KEY (lens_id) REFERENCES lenses (id);