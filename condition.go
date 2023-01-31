package main

import "time"

type Condition struct {
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	TimeRange *TimeRange `json:"timeRange"`
}

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}
