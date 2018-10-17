package utils

import (
	"net/url"
	"strings"
)

// IsValidURL checks if a given string is a valid URL.
func IsValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)

	if err != nil {
		return false
	}

	return true
}

// RemovePrefix strips a prefix from a string if it exists.
func RemovePrefix(value string, prefix string) string {
	if strings.HasPrefix(value, prefix) {
		return strings.TrimPrefix(value, prefix)
	}

	return value
}

// EnsureSuffix makes sure a string ends with a given suffix.
func EnsureSuffix(value string, suffix string) string {
	if !strings.HasSuffix(value, suffix) {
		return value + suffix
	}

	return value
}
