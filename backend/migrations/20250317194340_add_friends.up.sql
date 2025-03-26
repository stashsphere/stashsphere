CREATE TYPE friend_request_state AS ENUM('pending', 'accepted', 'rejected');

CREATE TABLE friend_requests (
    id text PRIMARY KEY,
    sender_id text NOT NULL,
    receiver_id text NOT NULL,
    state friend_request_state NOT NULL DEFAULT 'pending',
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE friendships (
    friend1_id text NOT NULL,
    friend2_id text NOT NULL,
    friend_request_id text NOT NULL,
    PRIMARY KEY (friend1_id, friend2_id),
    FOREIGN KEY (friend1_id) REFERENCES users(id),
    FOREIGN KEY (friend2_id) REFERENCES users(id),
    FOREIGN KEY (friend_request_id) REFERENCES friend_requests(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

