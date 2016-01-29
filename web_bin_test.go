package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinStatic(t *testing.T) {
	_, err := Asset("web-bin/index.html")
	assert.True(t, err == nil, "index.html must exists")
}
