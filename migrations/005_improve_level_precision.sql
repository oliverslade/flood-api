--
-- Migration 005: Convert level columns from real (float32) to double precision (float64)
-- This improves precision for level measurements and matches our domain models
--

ALTER TABLE public.rainfalls 
ALTER COLUMN level TYPE double precision;

ALTER TABLE public.riverlevels 
ALTER COLUMN level TYPE double precision;

ANALYZE public.rainfalls;
ANALYZE public.riverlevels;
