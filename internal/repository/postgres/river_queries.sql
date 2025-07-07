-- name: GetRiverReadings :many
-- Get river level readings sorted in chronological order with pagination
SELECT timestamp, level
FROM riverlevels
ORDER BY timestamp ASC
LIMIT $1 OFFSET $2;

-- name: GetRiverReadingsWithStartDate :many
-- Get river level readings from a start date sorted in chronological order with pagination
SELECT timestamp, level
FROM riverlevels
WHERE timestamp >= $1
ORDER BY timestamp ASC
LIMIT $2 OFFSET $3;

-- name: CountRiverReadings :one
-- Count total river level readings
SELECT COUNT(*) FROM riverlevels;

-- name: CountRiverReadingsWithStartDate :one
-- Count river level readings from a start date
SELECT COUNT(*) FROM riverlevels
WHERE timestamp >= $1;
