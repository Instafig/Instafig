package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenKey(t *testing.T) {
	filter := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		key := GenerateKey()
		assert.True(t, !filter[key], "should not gen old key")
		filter[key] = true
	}
}
