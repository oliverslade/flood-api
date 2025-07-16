package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type RiverReading struct {
	Timestamp time.Time
	Level     float64 // Level in metres
}

func (r RiverReading) MarshalJSON() ([]byte, error) {
	type Alias RiverReading
	return json.Marshal(&struct {
		Level string `json:"Level"`
		*Alias
	}{
		// Format Level to 3 decimals
		Level: fmt.Sprintf("%.3f", r.Level),
		Alias: (*Alias)(&r),
	})
}

func (r *RiverReading) UnmarshalJSON(data []byte) error {
	type Alias RiverReading
	aux := &struct {
		Level interface{} `json:"Level"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Level.(type) {
	case string:
		level, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		r.Level = level
	case float64:
		r.Level = v
	default:
		return fmt.Errorf("invalid type for Level field")
	}

	return nil
}

type RainfallReading struct {
	Timestamp   time.Time
	Level       float64 // Level in metres
	StationName string
}

func (r RainfallReading) MarshalJSON() ([]byte, error) {
	type Alias RainfallReading
	return json.Marshal(&struct {
		Level string `json:"Level"`
		*Alias
	}{
		// Format Level to 3 decimals
		Level: fmt.Sprintf("%.3f", r.Level),
		Alias: (*Alias)(&r),
	})
}

func (r *RainfallReading) UnmarshalJSON(data []byte) error {
	type Alias RainfallReading
	aux := &struct {
		Level interface{} `json:"Level"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Level.(type) {
	case string:
		level, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		r.Level = level
	case float64:
		r.Level = v
	default:
		return fmt.Errorf("invalid type for Level field")
	}

	return nil
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
