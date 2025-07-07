--
-- Migration 002: Performance optimizations for API operations
-- Optimizes only the two operations specified in the OpenAPI spec:
-- 1. GET /rainfall/{station} - Get rainfall readings for a station sorted chronologically
-- 2. GET /river - Get river level readings sorted chronologically
--

-- PHASE 1: Create indexes for the specific API operations

-- 1. Index for GET /rainfall/{station} - filtering by station + chronological sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rainfalls_station_timestamp 
ON public.rainfalls (stationid, timestamp);

-- 2. Index for GET /river - chronological sorting of river levels
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_riverlevels_timestamp 
ON public.riverlevels (timestamp);

-- PHASE 2: Update statistics for query planner
ANALYZE public.rainfalls;
ANALYZE public.riverlevels;
