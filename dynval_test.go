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
                (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
                5
             )`
	clientData := &ClientData{
		AppKey:     "app1",
		OSType:     "ios",
		OSVersion:  "9.3",
		AppVersion: "",
		Ip:         "14.32.123.23",
		Lang:       "",
	}
	assert.True(t, EvalDynValFromExpString(code, clientData) == 4)
}

func TestCondValuesExecute(t *testing.T) {
	code := `(cond-values
                (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1
                (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2
                (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
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
	assert.True(t, EvalDynValFromExpString(code, clientData) == 2)
}

func TestDynValToJson(t *testing.T) {
	env := NewGlisp()
	code := `(cond
                (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1
                (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2
                (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
                5
             )`
	dval := NewDynValFromString(code, env)
	data, _ := dval.ToJson()
	assert.True(t, data == `[{"symbol":"cond"},[{"symbol":"and"},[{"symbol":"="},{"symbol":"LANG"},"zh"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.5"]],1,[{"symbol":"and"},[{"symbol":"="},{"symbol":"LANG"},"zh"],[{"symbol":"or"},[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.5"]]],2,[{"symbol":"and"},[{"symbol":"not="},{"symbol":"LANG"},"zh"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.5"]],3,[{"symbol":"and"},[{"symbol":"not="},{"symbol":"LANG"},"zh"],[{"symbol":"or"},[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.5"]]],4,5]`)
}

func TestJsonToSexpString(t *testing.T) {
	expected_code := `(cond (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2 (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3 (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4 5)`
	json := `[{"symbol":"cond"},[{"symbol":"and"},[{"symbol":"="},{"symbol":"LANG"},"zh"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.5"]],1,[{"symbol":"and"},[{"symbol":"="},{"symbol":"LANG"},"zh"],[{"symbol":"or"},[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.5"]]],2,[{"symbol":"and"},[{"symbol":"not="},{"symbol":"LANG"},"zh"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.5"]],3,[{"symbol":"and"},[{"symbol":"not="},{"symbol":"LANG"},"zh"],[{"symbol":"or"},[{"symbol":"\u003c"},{"symbol":"APP_VERSION"},"1.3.1"],[{"symbol":"\u003e="},{"symbol":"APP_VERSION"},"1.5"]]],4,5]`
	code, _ := JsonToSexpString(json)
	assert.True(t, code == expected_code)
}

func TestCondValuesToJson(t *testing.T) {
	env := NewGlisp()
	code := `(cond-values
                (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1
                (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2
                (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3
                (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4
                5
             )`
	dval := NewDynValFromString(code, env)
	data, _ := dval.ToJson()
	expected_json := `{"cond-values":[{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":1},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":2},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":3},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":4}],"default-value":5}`
	assert.True(t, data == expected_json)
}

func TestJsonToCondValues(t *testing.T) {
	json := `{"cond-values":[{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":1},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":2},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":3},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":4}],"default-value":5}`
	expected_code := `(cond-values (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2 (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3 (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4 5)`

	data, _ := JsonToSexpString(json)
	assert.True(t, data == expected_code)
}
