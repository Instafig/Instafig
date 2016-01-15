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

func TestVersionCompareFunctions(t *testing.T) {
	env := NewGlisp()
	var ret glisp.Sexp
	ret, _ = env.EvalString(`(ver= "1.1.1" "1.1.1")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver= "1.1.1" "1.1.2")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(ver= "1.3" "1.3.0")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver< "1.1.1" "1.1.2")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver< "1.3.1" "1.3")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(ver> "1.1.1" "1.1.2")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(ver> "1.3.1" "1.3")`)
	assert.True(t, ret == glisp.SexpBool(true))
}
