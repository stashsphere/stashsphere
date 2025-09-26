CREATE TABLE cart_entries (
  user_id text NOT NULL,
  thing_id text NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, thing_id),
  FOREIGN KEY (thing_id) REFERENCES things(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE (user_id, thing_id)
);