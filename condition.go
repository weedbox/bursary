package main

import "time"

type Condition struct {
	Page      int          `json:"page"`
	Limit     int          `json:"limit"`
	TimeRange *TimeRange   `json:"timeRange,omitempty"`
	Sort      []*SortField `json:"sort,omitempty"`
}

type TimeRange struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type SortField struct {
	Field     string `json:"field"`
	Ascending bool   `json:"asc"`
}
