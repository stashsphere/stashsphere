CREATE TYPE sharing_state AS ENUM('private', 'friends', 'friends-of-friends');

ALTER TABLE things ADD COLUMN sharing_state sharing_state NOT NULL DEFAULT 'private';
ALTER TABLE lists  ADD COLUMN sharing_state sharing_state NOT NULL DEFAULT 'private';