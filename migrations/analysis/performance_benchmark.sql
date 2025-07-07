--
-- Performance benchmarking script
-- Tests the exact API query patterns from the OpenAPI specification
-- Run this before and after applying optimizations to measure improvements
--

-- Enable timing to see execution times
\timing on

-- Clear any cached plans
DISCARD PLANS;

-- ===========================================
-- API: GET /rainfall/{station}
-- ===========================================

-- Test 1: Basic rainfall query (default: page=1, pagesize=12)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level, stationid as station
FROM public.rainfalls 
WHERE stationid = '010660'  -- catcleugh (actual station ID)
ORDER BY timestamp 
LIMIT 12;

-- Test 2: Rainfall query with pagination (page=3, pagesize=12)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level, stationid as station
FROM public.rainfalls 
WHERE stationid = '010660'
ORDER BY timestamp 
LIMIT 12 OFFSET 24;  -- (page-1) * pagesize = (3-1) * 12 = 24

-- Test 3: Rainfall query with custom page size (pagesize=50)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level, stationid as station
FROM public.rainfalls 
WHERE stationid = '010660'
ORDER BY timestamp 
LIMIT 50;

-- Test 4: Rainfall query with start date filter
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level, stationid as station
FROM public.rainfalls 
WHERE stationid = '010660'
AND timestamp >= '2024-01-01 00:00:00'
ORDER BY timestamp 
LIMIT 12;

-- Test 5: Different station (haltwhistle)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level, stationid as station
FROM public.rainfalls 
WHERE stationid = '014555'  -- haltwhistle
ORDER BY timestamp 
LIMIT 12;

-- ===========================================
-- API: GET /river
-- ===========================================

-- Test 6: Basic river query (default: page=1, pagesize=12)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level
FROM public.riverlevels 
ORDER BY timestamp 
LIMIT 12;

-- Test 7: River query with pagination (page=10, pagesize=12)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level
FROM public.riverlevels 
ORDER BY timestamp 
LIMIT 12 OFFSET 108;  -- (page-1) * pagesize = (10-1) * 12 = 108

-- Test 8: River query with custom page size (pagesize=100)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level
FROM public.riverlevels 
ORDER BY timestamp 
LIMIT 100;

-- Test 9: River query with start date filter
EXPLAIN (ANALYZE, BUFFERS) 
SELECT timestamp, level
FROM public.riverlevels 
WHERE timestamp >= '2024-01-01 00:00:00'
ORDER BY timestamp 
LIMIT 12;
