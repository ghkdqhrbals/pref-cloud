package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type RequestMetrics struct {
	Duration    time.Duration
	TTFB        time.Duration
	Status      int
	StartedAt   time.Time
	CompletedAt time.Time
}

type PostData struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Results struct {
	URL              string                   `json:"url"`
	Method           string                   `json:"method"`
	TotalRequests    int                      `json:"total_requests"`
	TotalErrors      int                      `json:"total_errors"`
	TotalSuccess     int                      `json:"total_success"` // Computed field, consider omitting from JSON if not set explicitly
	StatusCodeCount  map[int]int              `json:"status_code_count"`
	TotalUsers       int                      `json:"total_users"`
	TotalDuration    time.Duration            `json:"total_duration"`    // May need custom marshaling
	MTTFBAverage     time.Duration            `json:"mttfb_average"`     // May need custom marshaling
	MTTFBPercentiles map[string]time.Duration `json:"mttfb_percentiles"` // May need custom marshaling
	TPSAverage       float64                  `json:"tps_average"`
	TPSPercentiles   map[string]float64       `json:"tps_percentiles"`
}

// DurationMS is init.sql custom type that embeds time.Duration for custom JSON marshaling.
type DurationMS time.Duration

// MarshalJSON converts the DurationMS to init.sql JSON object in milliseconds.
func (d DurationMS) MarshalJSON() ([]byte, error) {
	// Convert the duration to milliseconds and marshal to JSON number.
	ms := time.Duration(d).Milliseconds()
	return json.Marshal(ms)
}

// ResultsJSON is init.sql helper struct for JSON marshaling with duration in milliseconds.
type ResultsJSON struct {
	gorm.Model
	URL              string         `json:"url"`
	Method           string         `json:"method"`
	TotalRequests    int            `json:"total_requests"`
	TotalErrors      int            `json:"total_errors"`
	TotalSuccess     int            `json:"total_success"`
	StatusCodeCount  datatypes.JSON `gorm:"type:json"`
	TotalUsers       int            `json:"total_users"`
	TotalDuration    int64          `json:"total_duration"` // Milliseconds
	MTTFBAverage     string         `json:"mttfb_average"`  // Milliseconds
	MTTFBPercentiles datatypes.JSON `gorm:"type:json"`      // Milliseconds
	TPSAverage       float64        `json:"tps_average"`
	TPSPercentiles   datatypes.JSON `gorm:"type:json"`
}

func formatDuration(duration time.Duration) string {
	// Calculate milliseconds and microseconds
	milliseconds := duration.Milliseconds()
	microseconds := duration.Microseconds() % 1000
	// Keep microseconds with 4 decimal places
	microStr := fmt.Sprintf("%02d", microseconds)
	// Format the duration string with milliseconds and microseconds
	return fmt.Sprintf("%d.%sms", milliseconds, microStr)
}

// MarshalJSON customizes JSON output for Results.
func (r Results) MarshalJSON() ResultsJSON {

	// Convert MTTFB percentiles to milliseconds
	mttfbPercentiles := make(map[string]string)
	for k, v := range r.MTTFBPercentiles {
		mttfbPercentiles[k] = formatDuration(v) // Convert to seconds
	}

	mttfbPercentilesJSON, _ := json.Marshal(mttfbPercentiles)
	tpsPercentilesJSON, _ := json.Marshal(r.TPSPercentiles)
	statusCodeCountJSON, _ := json.Marshal(r.StatusCodeCount)

	resultsJSON := ResultsJSON{
		URL:              r.URL,
		Method:           r.Method,
		TotalRequests:    r.TotalRequests,
		TotalErrors:      r.TotalErrors,
		TotalSuccess:     r.TotalSuccess,
		StatusCodeCount:  statusCodeCountJSON,
		TotalUsers:       r.TotalUsers,
		TotalDuration:    r.TotalDuration.Milliseconds(),
		MTTFBAverage:     formatDuration(r.MTTFBAverage), // Convert to seconds
		MTTFBPercentiles: mttfbPercentilesJSON,
		TPSAverage:       r.TPSAverage,
		TPSPercentiles:   tpsPercentilesJSON,
	}

	return resultsJSON
}

func (r Results) String() string {
	return fmt.Sprintf(
		"Total Requests: %d\n"+
			"Total Errors: %d\n"+
			"Total Success: %d\n"+
			"Status Code Distribution: %v\n"+
			"Total Users: %d\n"+
			"Total Duration: %v\n"+
			"Mean Time To First Byte (MTTFB) Average: %v\n"+
			"MTTFB Percentiles: %v\n"+
			"Transactions Per Second (TPS) Average: %.2f\n"+
			"TPS Percentiles: %v\n",
		r.TotalRequests,
		r.TotalErrors,
		r.TotalSuccess,
		r.StatusCodeCount,
		r.TotalUsers,
		r.TotalDuration,
		r.MTTFBAverage,
		r.MTTFBPercentiles,
		r.TPSAverage,
		r.TPSPercentiles,
	)
}
