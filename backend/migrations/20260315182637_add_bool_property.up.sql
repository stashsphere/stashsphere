ALTER TYPE property_type ADD VALUE IF NOT EXISTS 'boolean' ;
COMMIT;
ALTER TABLE properties ADD COLUMN value_boolean boolean;
ALTER TABLE properties ADD CONSTRAINT chk_boolean_has_value CHECK (type != 'boolean' OR value_boolean IS NOT NULL);