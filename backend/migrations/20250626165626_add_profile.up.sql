CREATE TABLE profiles (
    id text PRIMARY KEY,
    full_name text NOT NULL,
    information text NOT NULL,
    image_id text,
    user_id text UNIQUE,
    FOREIGN KEY (image_id) REFERENCES images(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
