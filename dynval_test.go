package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhemao/glisp/interpreter"
	"testing"
)

func TestStub(t *testing.T) {
	assert.True(t, true, "This is good. Canary test passing")
}

func TestDynValTruncate(t *testing.T) {
	env := glisp.NewGlisp()
	dv := NewDynValFromString("(+ 1 2)(* 4 5)", env)
	assert.True(t, dv.Sexp_str == "(+ 1 2)")
}

func TestDynValExecute(t *testing.T) {
	code := `(cond (= LANG "zh")
			(cond (and (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 101)
			(cond (and (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 2 3))`
	clientData := &ClientData{
		AppKey:     "app1",
		OSType:     "ios",
		OSVersion:  "9.3",
		AppVersion: "1.1",
		Ip:         "14.32.123.23",
		Lang:       "zh",
	}
	assert.True(t, EvalDynVal(code, clientData) == 101)
}

func TestDynValToJson(t *testing.T) {
	env := glisp.NewGlisp()
	code := `(cond (= LANG "zh")
			(cond (and (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 101)
			(cond (and (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 2 3))`
	dval := NewDynValFromString(code, env)
	data, _ := dval.ToJson()
	assert.True(t, data == `[{"Symbol":"cond"},[{"Symbol":"="},{"Symbol":"LANG"},"zh"],[{"Symbol":"cond"},[{"Symbol":"and"},[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],1,101],[{"Symbol":"cond"},[{"Symbol":"and"},[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],2,3]]`)
}
