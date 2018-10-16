package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidURLWorksForValidURLs(t *testing.T) {
	assert.True(t, IsValidURL("http://example.com"))
	assert.True(t, IsValidURL("https://example.com"))
	assert.True(t, IsValidURL("http://example.com/some/path"))
	assert.True(t, IsValidURL("//example.com"))
}

func TestIsValidURLWorksForInvalidURLs(t *testing.T) {
	assert.False(t, IsValidURL("http//example.com"))
	assert.False(t, IsValidURL("some/string/like/a/path"))
	assert.False(t, IsValidURL("somewhere.com"))
}
