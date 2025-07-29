-- Test migration 004: Add NOT NULL constraints

ALTER TABLE rainfalls 
ALTER COLUMN timestamp SET NOT NULL,
ALTER COLUMN stationid SET NOT NULL,
ALTER COLUMN level SET NOT NULL;

ALTER TABLE riverlevels 
ALTER COLUMN timestamp SET NOT NULL,
ALTER COLUMN level SET NOT NULL;

ALTER TABLE stationnames 
ALTER COLUMN id SET NOT NULL,
ALTER COLUMN name SET NOT NULL;

ANALYZE rainfalls;
ANALYZE riverlevels;
ANALYZE stationnames; 