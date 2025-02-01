CREATE TABLE users (
  id text PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL
);
CREATE TABLE images (
  id text PRIMARY KEY,
  name text NOT NULL,
  mime text NOT NULL,
  hash text NOT NULL,
  owner_id text NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (owner_id) REFERENCES users(id)
);
CREATE TABLE things (
  id text PRIMARY KEY,
  name text NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  owner_id text NOT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id)
);
CREATE TABLE lists (
  id text PRIMARY KEY,
  name text NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  owner_id text NOT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id)
);
CREATE TABLE images_things (
  image_id text NOT NULL,
  thing_id text NOT NULL,
  PRIMARY KEY (image_id, thing_id),
  FOREIGN KEY (thing_id) REFERENCES images(id),
  FOREIGN KEY (image_id) REFERENCES things(id)
);
CREATE TABLE lists_things (
  list_id text NOT NULL,
  thing_id text NOT NULL,
  PRIMARY KEY (list_id, thing_id),
  FOREIGN KEY (list_id) REFERENCES lists(id),
  FOREIGN KEY (thing_id) REFERENCES things(id)
);
CREATE TYPE property_type AS ENUM('float', 'datetime', 'string');
CREATE TABLE properties (
  id text PRIMARY KEY,
  type property_type NOT NULL,
  name VARCHAR(255) NOT NULL,
  
  value_string VARCHAR(255),
  value_datetime TIMESTAMP,
  value_float FLOAT,
  
  unit VARCHAR(20),
  
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  thing_id text NOT NULL,
  FOREIGN KEY (thing_id) REFERENCES things(id),
  UNIQUE (name, thing_id),
  
  CONSTRAINT chk_string_has_value CHECK (type != 'string' OR value_string IS NOT NULL),
  CONSTRAINT chk_datetime_has_value CHECK (type != 'datetime' OR value_datetime IS NOT NULL),
  CONSTRAINT chk_float_has_value CHECK (type != 'float' OR value_float IS NOT NULL)
);