package controllers

import (
	"fmt"
	"strconv"
	"time"
)

func parseLimit(raw string, def int, max int) int {
	if raw == "" {
		return def
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return def
	}
	if value > max {
		return max
	}
	return value
}

func parseOffset(raw string) int {
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func parseTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp %q: must use RFC3339 format", raw)
	}
	return &parsed, nil
}