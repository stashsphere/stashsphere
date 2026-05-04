CREATE TYPE export_status AS ENUM ('pending', 'processing', 'done', 'error');

CREATE TABLE exports (
    id          TEXT            NOT NULL PRIMARY KEY,
    owner_id    TEXT            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      export_status   NOT NULL DEFAULT 'pending',
    file_path   TEXT,
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at  TIMESTAMP,
    error_msg   TEXT
);

CREATE UNIQUE INDEX exports_user_active
  ON exports (owner_id)
  WHERE status IN ('pending', 'processing');
