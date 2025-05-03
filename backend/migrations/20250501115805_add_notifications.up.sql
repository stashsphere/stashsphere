CREATE TABLE notifications (
    id text PRIMARY KEY,
    recipient_id text NOT NULL,
    FOREIGN KEY (recipient_id) REFERENCES users(id),
    content_type VARCHAR(100) NOT NULL,
    content JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP
);