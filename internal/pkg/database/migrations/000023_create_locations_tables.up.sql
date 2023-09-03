CREATE SCHEMA locations;

-- importers are the names of endpoints which process locations data
CREATE TABLE locations.importers (
  id SERIAL NOT NULL PRIMARY KEY,
  name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (name)
);
CREATE TRIGGER set_timestamp
  BEFORE UPDATE ON locations.importers
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();

-- callers are the names of devices or programs which invoke importers
CREATE TABLE locations.callers (
  id SERIAL NOT NULL PRIMARY KEY,
  name text CONSTRAINT name_present CHECK ((name != '') IS TRUE),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (name)
);
CREATE TRIGGER set_timestamp
  BEFORE UPDATE ON locations.callers
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();

-- reasons are strings which explain why a location point was logged
CREATE TABLE locations.reasons (
 id SERIAL NOT NULL PRIMARY KEY,
 reference text CONSTRAINT reference_present CHECK ((reference != '') IS TRUE),
 UNIQUE (reference)
);

-- activities are events spanning a period of time, they may have related location data
CREATE TABLE locations.activities (
  id SERIAL NOT NULL PRIMARY KEY,
  title text CONSTRAINT title_present CHECK ((title != '') IS TRUE),
  description text NOT NULL DEFAULT '',

  start_time TIMESTAMP NOT NULL DEFAULT NOW(),
  end_time TIMESTAMP NOT NULL DEFAULT NOW(),

  importer_id INTEGER REFERENCES locations.importers(id) DEFAULT NULL,
  caller_id INTEGER REFERENCES locations.callers(id) DEFAULT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_timestamp
  BEFORE UPDATE ON locations.activities
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();

-- points are records of a location for a given timestamp
CREATE TABLE locations.points (
  id SERIAL NOT NULL PRIMARY KEY,

  latitude float NOT NULL DEFAULT 0,
  longitude float NOT NULL DEFAULT 0,
  altitude float NOT NULL DEFAULT 0,

  accuracy float NOT NULL DEFAULT 0,
  vertical_accuracy float NOT NULL DEFAULT 0,

  velocity float NOT NULL DEFAULT 0,

  was_offline bool NOT NULL DEFAULT FALSE,

  importer_id INTEGER REFERENCES locations.importers(id) NOT NULL,
  caller_id INTEGER REFERENCES locations.callers(id) NOT NULL,
  reason_id INTEGER REFERENCES locations.reasons(id) NOT NULL,

  -- activity_id is only set when the point was created at the same time as the activity
  activity_id INTEGER REFERENCES locations.activities(id) DEFAULT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_timestamp
  BEFORE UPDATE ON locations.points
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();
