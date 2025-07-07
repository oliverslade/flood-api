--
-- Migration 004: Add NOT NULL constraints to columns that should not be null
-- This migration adds NOT NULL constraints to improve type safety and allow
-- sqlc to generate cleaner types without sql.Null* wrappers
--

ALTER TABLE public.rainfalls 
ALTER COLUMN timestamp SET NOT NULL,
ALTER COLUMN stationid SET NOT NULL,
ALTER COLUMN level SET NOT NULL;

ALTER TABLE public.riverlevels 
ALTER COLUMN timestamp SET NOT NULL,
ALTER COLUMN level SET NOT NULL;

ALTER TABLE public.stationnames 
ALTER COLUMN id SET NOT NULL,
ALTER COLUMN name SET NOT NULL;

ANALYZE public.rainfalls;
ANALYZE public.riverlevels;
ANALYZE public.stationnames;
