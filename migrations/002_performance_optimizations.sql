--
-- Migration 002: Performance optimizations for API operations
-- Optimizes only the two operations specified in the OpenAPI spec:

-- Index for GET /rainfall/{station} - filtering by station + chronological sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rainfalls_station_timestamp 
ON public.rainfalls (stationid, timestamp);

-- Index for GET /river - chronological sorting of river levels
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_riverlevels_timestamp 
ON public.riverlevels (timestamp);

ANALYZE public.rainfalls;
ANALYZE public.riverlevels;
