package utils

import (
	"strconv"
)

// ParseIntOrDefault parses string to int, fallback to default on error
func ParseIntOrDefault(s string, defaultValue int) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
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
