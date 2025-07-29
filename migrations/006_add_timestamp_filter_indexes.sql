-- 
-- Migration 006: Add composite indexes for timestamp filtering
-- Optimize queries that filter by timestamp AND paginate
--

-- Composite index for river readings with timestamp filter
-- This speeds up queries like: WHERE timestamp >= $1 ORDER BY timestamp ASC
CREATE INDEX IF NOT EXISTS idx_riverlevels_timestamp_desc 
ON public.riverlevels (timestamp DESC);

-- Composite index for rainfall readings with station AND timestamp filter  
-- This speeds up queries like: WHERE stationid = $1 AND timestamp >= $2 ORDER BY timestamp ASC
CREATE INDEX IF NOT EXISTS idx_rainfalls_station_timestamp_desc
ON public.rainfalls (stationid, timestamp DESC);

ANALYZE public.riverlevels;
ANALYZE public.rainfalls; 