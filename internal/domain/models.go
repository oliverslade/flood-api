package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrNotFound = errors.New("not found")

const tsLayout = "2006-01-02T15:04:05" // spec‑defined timestamp layout

func formatTS(t time.Time) string         { return t.Format(tsLayout) }
func parseTS(s string) (time.Time, error) { return time.Parse(tsLayout, s) }

func roundToDecimalPlaces(v float64, dp int) float64 {
	m := math.Pow10(dp)
	return math.Round(v*m) / m
}

type RiverReading struct {
	Timestamp time.Time `json:"timestamp"`
	Level     float64   `json:"level"`
}

func (r RiverReading) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp string  `json:"timestamp"`
		Level     float64 `json:"level"`
	}{
		Timestamp: formatTS(r.Timestamp),
		Level:     roundToDecimalPlaces(r.Level, 3),
	})
}

type RainfallReading struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       float64   `json:"level"`
	StationName string    `json:"station"`
}

func (r RainfallReading) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp string  `json:"timestamp"`
		Level     float64 `json:"level"`
		Station   string  `json:"station"`
	}{
		Timestamp: formatTS(r.Timestamp),
		Level:     roundToDecimalPlaces(r.Level, 3),
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

var (
	_ json.Marshaler = (*RiverReading)(nil)
	_ json.Marshaler = (*RainfallReading)(nil)
)

func ParseTimestamp(s string) (time.Time, error) {
	t, err := parseTS(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp: %w", err)
	}
	return t, nil
}
