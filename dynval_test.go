package main

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestDynValTruncate(t *testing.T) {
	env := NewGlisp()
	dv := NewDynValFromString("(+ 1 2)(* 4 5)", env)
	assert.True(t, dv.Sexp_str == "(+ 1 2)")
}

func TestDynValExecute(t *testing.T) {
	code := `(cond
                (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1
                (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2
                (and (!= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (!= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
                5
             )`
	clientData := &ClientData{
		AppKey:     "app1",
		OSType:     "ios",
		OSVersion:  "9.3",
		AppVersion: "1.1",
		Ip:         "14.32.123.23",
		Lang:       "zh",
	}
	assert.True(t, EvalDynVal(code, clientData) == 2)
}

func TestDynValToJson(t *testing.T) {
	env := NewGlisp()
	code := `(cond
                (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1
                (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2
                (and (!= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (!= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
                5
             )`
	dval := NewDynValFromString(code, env)
	data, _ := dval.ToJson()
	assert.True(t, data == `[{"Symbol":"cond"},[{"Symbol":"and"},[{"Symbol":"="},{"Symbol":"LANG"},"zh"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],1,[{"Symbol":"and"},[{"Symbol":"="},{"Symbol":"LANG"},"zh"],[{"Symbol":"or"},[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.5"]]],2,[{"Symbol":"and"},[{"Symbol":"!="},{"Symbol":"LANG"},"zh"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],3,[{"Symbol":"and"},[{"Symbol":"!="},{"Symbol":"LANG"},"zh"],[{"Symbol":"or"},[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.5"]]],4,5]`)
}

func TestJsonToSexpString(t *testing.T) {
	excepted_code := `(cond (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2 (and (!= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3 (and (!= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4 5)`
	json := `[{"Symbol":"cond"},[{"Symbol":"and"},[{"Symbol":"="},{"Symbol":"LANG"},"zh"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],1,[{"Symbol":"and"},[{"Symbol":"="},{"Symbol":"LANG"},"zh"],[{"Symbol":"or"},[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.5"]]],2,[{"Symbol":"and"},[{"Symbol":"!="},{"Symbol":"LANG"},"zh"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.5"]],3,[{"Symbol":"and"},[{"Symbol":"!="},{"Symbol":"LANG"},"zh"],[{"Symbol":"or"},[{"Symbol":"\u003c"},{"Symbol":"APP_VERSION"},"1.3.1"],[{"Symbol":"\u003e="},{"Symbol":"APP_VERSION"},"1.5"]]],4,5]`
	code, _ := JsonToSexpString(json)
	assert.True(t, code == excepted_code)
}
