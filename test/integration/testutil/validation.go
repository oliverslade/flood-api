package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	timestampPattern = regexp.MustCompile(`^2[0-9]{3}-(0[0-9]|1[0-2])-([0-2][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z?$`)
	timeLayout       = "2006-01-02T15:04:05Z"
)

// APIResponse represents the standard readings response
type APIResponse struct {
	Readings []Reading `json:"readings"`
}

// Reading represents a measurement reading
type Reading struct {
	Timestamp string  `json:"timestamp"`
	Level     float64 `json:"level"`
	Station   string  `json:"station,omitempty"`
}

var HTTPClient = &http.Client{
	Timeout: 0, // Let context handle timeouts
}

// MustGET performs a GET request and returns parsed response
func MustGET(tb testing.TB, ctx context.Context, url string) APIResponse {
	tb.Helper()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(tb, err)

	resp, err := HTTPClient.Do(req)
	require.NoError(tb, err)
	defer resp.Body.Close()

	require.Equal(tb, http.StatusOK, resp.StatusCode, "URL: %s", url)
	assert.Equal(tb, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(tb, err)

	var result APIResponse
	err = json.Unmarshal(body, &result)
	require.NoError(tb, err)

	return result
}

// ExpectHTTPError performs a GET and expects a specific status code
func ExpectHTTPError(tb testing.TB, ctx context.Context, url string, expectedStatus int) {
	tb.Helper()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(tb, err)

	resp, err := HTTPClient.Do(req)
	require.NoError(tb, err)
	defer resp.Body.Close()

	assert.Equal(tb, expectedStatus, resp.StatusCode, "URL: %s", url)
}

// ValidateReading checks a reading matches OpenAPI spec constraints
func ValidateReading(tb testing.TB, r Reading, expectedStation string) {
	tb.Helper()

	// Validate timestamp format
	assert.Regexp(tb, timestampPattern, r.Timestamp, "Invalid timestamp format")

	// Parse to ensure it's valid
	_, err := time.Parse(timeLayout, r.Timestamp)
	if err != nil {
		// Try without Z suffix for backwards compat
		_, err = time.Parse("2006-01-02T15:04:05", r.Timestamp)
	}
	require.NoError(tb, err, "Failed to parse timestamp: %s", r.Timestamp)

	// Level must be non-negative per OpenAPI spec
	assert.GreaterOrEqual(tb, r.Level, 0.0)

	// Station validation if expected
	if expectedStation != "" {
		assert.Equal(tb, expectedStation, r.Station)
	}
}

// AssertReadingsEqual compares actual vs expected readings
func AssertReadingsEqual(tb testing.TB, expected []Reading, actual []Reading) {
	tb.Helper()

	require.Equal(tb, len(expected), len(actual), "Number of readings mismatch")

	for i, exp := range expected {
		act := actual[i]
		assert.Equal(tb, exp.Timestamp, act.Timestamp, "Reading %d timestamp", i)
		assert.InDelta(tb, exp.Level, act.Level, 0.001, "Reading %d level", i) // 3 decimal places
		if exp.Station != "" {
			assert.Equal(tb, exp.Station, act.Station, "Reading %d station", i)
		}
	}
}
