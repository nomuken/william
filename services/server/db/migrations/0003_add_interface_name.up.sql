ALTER TABLE interfaces ADD COLUMN name TEXT NOT NULL DEFAULT '';

UPDATE interfaces
SET name = id
WHERE name = '';
