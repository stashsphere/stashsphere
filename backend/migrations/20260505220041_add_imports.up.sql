CREATE TYPE import_status AS ENUM ('pending', 'processing', 'done', 'error');

CREATE TABLE imports (
    id               TEXT           NOT NULL PRIMARY KEY,
    owner_id         TEXT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status           import_status  NOT NULL DEFAULT 'pending',
    file_path        TEXT,
    created_at       TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at     TIMESTAMP,
    error_msg        TEXT,
    things_imported  INT,
    lists_imported   INT,
    images_imported  INT
);

CREATE UNIQUE INDEX imports_user_active ON imports (owner_id) WHERE status IN ('pending', 'processing');
