package domain

import (
	"encoding/json"
	"errors"
	"math"
	"time"
)

var ErrNotFound = errors.New("not found")

// roundLevel rounds to 3 decimal places as per API spec
func roundLevel(v float64) float64 {
	return math.Round(v*1000) / 1000
}

type RiverReading struct {
	Timestamp time.Time `json:"timestamp"`
	Level     float64   `json:"level"`
}

func (r RiverReading) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp time.Time `json:"timestamp"`
		Level     float64   `json:"level"`
	}{
		Timestamp: r.Timestamp,
		Level:     roundLevel(r.Level),
	})
}

type RainfallReading struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       float64   `json:"level"`
	StationName string    `json:"station"`
}

func (r RainfallReading) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp time.Time `json:"timestamp"`
		Level     float64   `json:"level"`
		Station   string    `json:"station"`
	}{
		Timestamp: r.Timestamp,
		Level:     roundLevel(r.Level),
		Station:   r.StationName,
	})
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
