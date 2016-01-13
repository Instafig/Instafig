package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhemao/glisp/interpreter"
	"testing"
)

func TestStub(t *testing.T) {
	assert.True(t, true, "This is good. Canary test passing")
}

func TestDvalTruncate(t *testing.T) {
	env := glisp.NewGlisp()
	dv := NewDynValFromString("(+ 1 2)(* 4 5)", env)
	assert.True(t, dv.Sexp_str == "(+ 1 2)")
}
