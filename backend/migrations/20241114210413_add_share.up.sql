CREATE TABLE shares (
  id text PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  target_user_id text NOT NULL,
  owner_id text NOT NULL,
  FOREIGN KEY (target_user_id) REFERENCES users(id),
  FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE shares_things (
    share_id text NOT NULL,
    thing_id text NOT NULL,
    PRIMARY KEY (share_id, thing_id),
    FOREIGN KEY (thing_id) REFERENCES things(id),
    FOREIGN KEY (share_id) REFERENCES shares(id)
);

CREATE TABLE shares_lists (
    share_id text NOT NULL,
    list_id text NOT NULL,
    PRIMARY KEY (share_id, list_id),
    FOREIGN KEY (list_id) REFERENCES lists(id),
    FOREIGN KEY (share_id) REFERENCES shares(id)
);