-- Test migration 003: Convert timestamp columns from text to timestamp

ALTER TABLE rainfalls 
ADD COLUMN timestamp_new timestamp;

ALTER TABLE riverlevels 
ADD COLUMN timestamp_new timestamp;

UPDATE rainfalls 
SET timestamp_new = timestamp::timestamp
WHERE timestamp IS NOT NULL AND timestamp != '';

UPDATE riverlevels 
SET timestamp_new = timestamp::timestamp
WHERE timestamp IS NOT NULL AND timestamp != '';

ALTER TABLE rainfalls 
DROP COLUMN timestamp;

ALTER TABLE rainfalls 
RENAME COLUMN timestamp_new TO timestamp;

ALTER TABLE riverlevels 
DROP COLUMN timestamp;

ALTER TABLE riverlevels 
RENAME COLUMN timestamp_new TO timestamp;

ANALYZE rainfalls;
ANALYZE riverlevels; 