package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhemao/glisp/interpreter"

	"testing"
)

func TestMacroCondValues(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

	ret, _ := env.EvalString("(cond-values false 1 true 2 3)")
	assert.True(t, ret == glisp.SexpInt(2))
}

func TestVersionCompareFunctions(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

	var ret glisp.Sexp
	ret, _ = env.EvalString(`(ver= "1.1.1" "1.1.1")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver= "1.1.1" "1.1.2")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(ver= "1.3" "1.3.0")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver< "1.4.1" "1.14.2")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver< "1.3.1" "1.3")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(ver> "1.12.1" "1.2.2")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver> "1.3.1" "1.3")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver> "1.3.3333" "1.3.2222")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(ver!= "1.1.1" "1.1.1")`)
	assert.True(t, ret == glisp.SexpBool(false))

}

func TestStringContainsFunctions(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

	var ret glisp.Sexp
	ret, _ = env.EvalString(`(str-contains? "abc" "abc")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-contains? "abc" "ayz")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(str-contains? "_abc_" "abc")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-not-contains? "abc" "bb")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-not-contains? "abc" "abcd")`)
	assert.True(t, ret == glisp.SexpBool(true))
}

func TestStringWildcardMatchFunctions(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

	var ret glisp.Sexp
	ret, _ = env.EvalString(`(str-wcmatch? "axyzc" "a*c")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-wcmatch? "_axyzc" "a*c")`)
	assert.True(t, ret == glisp.SexpBool(false))
	ret, _ = env.EvalString(`(str-wcmatch? "abc" "a?c")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-not-wcmatch? "abbc" "a?c")`)
	assert.True(t, ret == glisp.SexpBool(true))
	ret, _ = env.EvalString(`(str-not-wcmatch? "a2c" "a\\dc")`)
	assert.True(t, ret == glisp.SexpBool(true))
}
