package utils

import (
	"strconv"
)

// ParseIntOrDefault parses string to int, fallback to default on error
func ParseIntOrDefault(s string, defaultVal int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultVal
}

// ParseFloatOrDefault parses string to float64, fallback to default on error
func ParseFloatOrDefault(s string, defaultVal float64) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return defaultVal
}

// ParseBoolPointer parses "1" or "0" into *bool, returns nil if invalid or empty
func ParseBoolPointer(s string) *bool {
	if s == "1" {
		b := true
		return &b
	}
	if s == "0" {
		b := false
		return &b
	}
	return nil
}
