package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynValTruncate(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

	dv := NewDynValFromString("(+ 1 2)(* 4 5)", env)
	assert.True(t, dv.SexpStr == "(+ 1 2)")
}

func TestDynValExecute(t *testing.T) {
	code := `(cond
                (and (str= LANG "zh") (ver>= APP_VERSION "1.3.1") (ver< APP_VERSION "1.5")) 1
                (and (str= LANG "zh") (or (ver< APP_VERSION "1.3.1") (ver>= APP_VERSION "1.5"))) 2
                (and (str!= LANG "zh") (ver>= APP_VERSION "1.3.1") (ver< APP_VERSION "1.5")) 3
                (and (str!= LANG "zh") (or (ver< APP_VERSION "1.3.1") (ver>= APP_VERSION "1.5"))) 4
                5
             )`
	clientData := &ClientData{
		AppKey:     "app1",
		OSType:     "ios",
		OSVersion:  "9.3",
		AppVersion: "1.5",
		Ip:         "14.32.123.23",
		Lang:       "",
	}
	res, err := EvalDynVal(NewDynValFromSexpStringDefault(code), clientData)
	assert.True(t, err == nil)
	assert.True(t, res == 4)
}

func TestCondValuesExecute(t *testing.T) {
	code := `(cond-values
                (and (str= LANG "zh") (ver>= APP_VERSION "1.3.1") (ver< APP_VERSION "1.5")) 1
                (and (str= LANG "zh") (or (ver< APP_VERSION "1.3.1") (ver>= APP_VERSION "1.5"))) 2
                (and (str!= LANG "zh") (ver>= APP_VERSION "1.3.1") (ver< APP_VERSION "1.5")) 3
                (and (str!= LANG "zh") (or (ver< APP_VERSION "1.3.1") (ver>= APP_VERSION "1.5"))) 4
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
	res, err := EvalDynVal(NewDynValFromSexpStringDefault(code), clientData)
	assert.True(t, err == nil)
	assert.True(t, res == 2)
}

func TestDynValToJson(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

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
	code, err := JsonToSexpString(json)
	assert.True(t, err == nil)
	assert.True(t, code == expected_code)
}

func TestCondValuesToJson(t *testing.T) {
	env := getGLispEnv()
	defer putGLispEnv(env)

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
	//json := `{"cond-values":[{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":1},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":2},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003e="},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003c"}],"func":"and"},"value":3},{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"zh"],"func":"not="},{"arguments":[{"arguments":[{"symbol":"APP_VERSION"},"1.3.1"],"func":"\u003c"},{"arguments":[{"symbol":"APP_VERSION"},"1.5"],"func":"\u003e="}],"func":"or"}],"func":"and"},"value":4}],"default-value":5}`
	//expected_code := `(cond-values (and (= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 1 (and (= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 2 (and (not= LANG "zh") (>= APP_VERSION "1.3.1") (< APP_VERSION "1.5")) 3 (and (not= LANG "zh") (or (< APP_VERSION "1.3.1") (>= APP_VERSION "1.5"))) 4 5)`
	json := `{"cond-values":[{"condition":{"arguments":[{"arguments":[{"symbol":"LANG"},"en"],"func":"str!="},{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver="},{"arguments":[{"symbol":"OS_VERSION"},"2.4.5"],"func":"ver<"}],"func":"or"},"value":123}],"default-value":456}`
	expected_code := `(cond-values (or (str!= LANG "en") (ver= APP_VERSION "1.0") (ver< OS_VERSION "2.4.5")) 123 456)`

	data, err := JsonToSexpString(json)
	assert.True(t, err == nil)
	assert.True(t, data == expected_code)
}

func TestVersionCondConfigValue(t *testing.T) {
	// bad json
	json := `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) == nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) == nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"}],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION", "1.0", "1.2"}],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.1.1.2.3.4"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"LANG"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"IP"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"OS_TYPE"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"DEVICE_ID"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"NETWORK"},"1.1"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	assert.True(t, CheckJsonString(json) != nil)

	// good json
	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver="},"value":"0"}],"default-value":"1"}`
	sep, _ := JsonToSexpString(json)
	dynval := NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.0"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.1"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "10.1"}) == "1")

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver>"},"value":"0"}],"default-value":"1"}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.0"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "10.1"}) == "0")

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver>="},"value":"0"}],"default-value":"1"}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.0.1"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.9.9"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.0"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "10.1.1"}) == "0")

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver<"},"value":"0"}],"default-value":"1"}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.0.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.9.9"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.0"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.1"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "10.1.1"}) == "1")

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"ver<="},"value":"0"}],"default-value":"1"}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.0.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "0.9.9"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.0"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "1.1"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{AppVersion: "10.1.1"}) == "1")

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"OS_VERSION"},"1.0"],"func":"ver!="},"value":"0"}],"default-value":"1"}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{OSVersion: "0.0.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{OSVersion: "0.9.9"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{OSVersion: "1.0"}) == "1")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{OSVersion: "1.1"}) == "0")
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{OSVersion: "10.1.1"}) == "0")
}

func TestStrCondConfigValue(t *testing.T) {
	//bad json
	json := `{"cond-values":[{"condition":{"arguments":[{"symbol":"OS_VERSION"},"1.0"],"func":"str="},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"APP_VERSION"},"1.0"],"func":"str="},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"LANG"},"en", "da"],"func":"str="},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"IP"}],"func":"str="},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"IP"}, "adsfa"],"func":"str-not-empty?"},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"IP"}, "adsfa"],"func":"str-empty?"},"value":0}],"default-value":1}`
	assert.True(t, CheckJsonString(json) != nil)

	// good json
	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"LANG"},"en"],"func":"str="},"value":0}],"default-value":1}`
	sep, _ := JsonToSexpString(json)
	dynval := NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{Lang: "en"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{Lang: "zh"}) == 1)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"LANG"},"en"],"func":"str!="},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{Lang: "en"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{Lang: "zh"}) == 0)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}],"func":"str-empty?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: ""}) == 0)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}],"func":"str-not-empty?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: ""}) == 1)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}, "beijing"],"func":"str-contains?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing,shanghai"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "tianjin,beijing,shanghai"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "hokong"}) == 1)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}, "beijing"],"func":"str-not-contains?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing,shanghai"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "tianjin,beijing,shanghai"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "hokong"}) == 0)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}, "*beijin*"],"func":"str-wcmatch?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing,shanghai"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "tianjin,beijing,shanghai"}) == 0)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "hokong"}) == 1)

	json = `{"cond-values":[{"condition":{"arguments":[{"symbol":"TIMEZONE"}, "*beijin*"],"func":"str-not-wcmatch?"},"value":0}],"default-value":1}`
	sep, _ = JsonToSexpString(json)
	dynval = NewDynValFromSexpStringDefault(sep)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "beijing,shanghai"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "tianjin,beijing,shanghai"}) == 1)
	assert.True(t, EvalDynValNoErr(dynval, &ClientData{TimeZone: "hokong"}) == 0)

	// todo: more wild-card match case
}
