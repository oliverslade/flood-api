-- Test migration 005: Convert level columns from real to double precision

ALTER TABLE rainfalls 
ALTER COLUMN level TYPE double precision;

ALTER TABLE riverlevels 
ALTER COLUMN level TYPE double precision;

ANALYZE rainfalls;
ANALYZE riverlevels; 