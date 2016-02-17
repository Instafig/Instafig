package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinStatic(t *testing.T) {
	_, err := Asset("web/index.html")
	assert.True(t, err == nil, "index.html must exists")

	_, err = Asset("web/non-exist-web-data-blabla.html")
	assert.True(t, err != nil, "must not exists")
}
