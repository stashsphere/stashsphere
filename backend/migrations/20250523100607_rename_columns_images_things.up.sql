ALTER TABLE images_things RENAME COLUMN image_id TO iimage_id;
ALTER TABLE images_things RENAME COLUMN thing_id TO image_id;
ALTER TABLE images_things RENAME COLUMN iimage_id TO thing_id;

ALTER TABLE images_things DROP CONSTRAINT images_things_image_id_fkey,
                          DROP CONSTRAINT images_things_thing_id_fkey;
ALTER TABLE images_things
ADD CONSTRAINT images_things_image_id_fkey FOREIGN KEY (image_id) REFERENCES images(id),
ADD CONSTRAINT images_things_thing_id_fkey FOREIGN KEY (thing_id) REFERENCES things(id);