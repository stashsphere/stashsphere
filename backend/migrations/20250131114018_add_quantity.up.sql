CREATE TABLE quantity_entries ( 
    id text PRIMARY KEY,   
    thing_id text NOT NULL,
    delta_value bigint NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (thing_id) REFERENCES things(id)
);

ALTER TABLE things ADD COLUMN quantity_unit text NOT NULL default 'pcs';