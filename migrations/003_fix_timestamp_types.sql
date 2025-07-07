--
-- Migration 003: Convert timestamp columns from text to timestamp
-- This migration converts string timestamps to proper PostgreSQL timestamp types
-- for better performance and to allow sqlc to scan directly into time.Time
--

ALTER TABLE public.rainfalls 
ADD COLUMN timestamp_new timestamp;

ALTER TABLE public.riverlevels 
ADD COLUMN timestamp_new timestamp;

UPDATE public.rainfalls 
SET timestamp_new = timestamp::timestamp
WHERE timestamp IS NOT NULL AND timestamp != '';

UPDATE public.riverlevels 
SET timestamp_new = timestamp::timestamp
WHERE timestamp IS NOT NULL AND timestamp != '';

ALTER TABLE public.rainfalls 
DROP COLUMN timestamp;

ALTER TABLE public.rainfalls 
RENAME COLUMN timestamp_new TO timestamp;

ALTER TABLE public.riverlevels 
DROP COLUMN timestamp;

ALTER TABLE public.riverlevels 
RENAME COLUMN timestamp_new TO timestamp;

ANALYZE public.rainfalls;
ANALYZE public.riverlevels;
