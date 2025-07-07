package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type RiverReading struct {
	Timestamp time.Time
	Level     float64 // Level in metres
}

type RainfallReading struct {
	Timestamp   time.Time
	Level       float64 // Level in metres
	StationName string
}

type Station struct {
	ID   string
	Name string
}

type PaginationParams struct {
	Page     int
	PageSize int
}

type GetReadingsParams struct {
	Pagination PaginationParams
	StartDate  *time.Time // Optional start date filter
}

type GetRainfallParams struct {
	StationName string
	GetReadingsParams
}
