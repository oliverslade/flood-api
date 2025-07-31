-- Test migration 001: Initial schema (clean version)

CREATE TABLE rainfalls (
    stationid text,
    "timestamp" text,
    level real
);

CREATE TABLE riverlevels (
    "timestamp" text,
    level real
);

CREATE TABLE stationnames (
    id text,
    name text
); 