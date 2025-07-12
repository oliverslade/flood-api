-- Test migration 002: Performance optimizations (clean version)

CREATE INDEX IF NOT EXISTS idx_rainfalls_station_timestamp 
ON rainfalls (stationid, timestamp);

CREATE INDEX IF NOT EXISTS idx_riverlevels_timestamp 
ON riverlevels (timestamp);

ANALYZE rainfalls;
ANALYZE riverlevels; 