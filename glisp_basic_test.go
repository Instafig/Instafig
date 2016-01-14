package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhemao/glisp/interpreter"

	"testing"
)

func TestMacroCondValues(t *testing.T) {
	env := NewGlisp()
	ret, _ := env.EvalString("(cond-values false 1 true 2 3)")
	assert.True(t, ret == glisp.SexpInt(2))
}
