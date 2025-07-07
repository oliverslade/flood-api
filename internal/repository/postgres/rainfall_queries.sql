-- name: GetRainfallReadingsByStation :many
-- Get rainfall readings for a station sorted in chronological order with pagination
SELECT timestamp, level, stationid
FROM rainfalls
WHERE stationid = $1
ORDER BY timestamp ASC
LIMIT $2 OFFSET $3;

-- name: GetRainfallReadingsByStationWithStartDate :many
-- Get rainfall readings for a station from a start date sorted in chronological order with pagination
SELECT timestamp, level, stationid
FROM rainfalls
WHERE stationid = $1 AND timestamp >= $2
ORDER BY timestamp ASC
LIMIT $3 OFFSET $4;

-- name: CountRainfallReadingsByStation :one
-- Count rainfall readings for a station
SELECT COUNT(*) FROM rainfalls
WHERE stationid = $1;

-- name: CountRainfallReadingsByStationWithStartDate :one
-- Count rainfall readings for a station from a start date
SELECT COUNT(*) FROM rainfalls
WHERE stationid = $1 AND timestamp >= $2;

-- name: GetStationByID :one
-- Get station information by ID for validation
SELECT id, name FROM stationnames
WHERE id = $1;

-- name: GetStationByName :one
-- Get station information by name for API lookups
SELECT id, name FROM stationnames
WHERE name = $1;
