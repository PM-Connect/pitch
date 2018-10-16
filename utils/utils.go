package utils

import "net/url"

// IsValidURL checks if a given string is a valid URL.
func IsValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)

	if err != nil {
		return false
	}

	return true
}
