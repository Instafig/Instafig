package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetServiceStatus(t *testing.T) {
	c := &gin.Context{}
	setServiceStatus(c, true)
	assert.True(t, getServiceStatus(c))

	setServiceStatus(c, false)
	assert.True(t, !getServiceStatus(c))
}
