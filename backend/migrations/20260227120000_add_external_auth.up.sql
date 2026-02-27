CREATE TABLE external_auth (
  user_id TEXT NOT NULL,
  provider VARCHAR(255) NOT NULL,
  subject VARCHAR(255) NOT NULL,
  PRIMARY KEY (user_id, provider),
  UNIQUE (provider, subject),
  FOREIGN KEY (user_id) REFERENCES users(id)
);

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
