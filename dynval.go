package main

import (
	"encoding/json"
	"github.com/zhemao/glisp/interpreter"
	"strconv"
	"strings"
)

type DynVal struct {
	Sexp     glisp.Sexp
	Sexp_str string
}

type ExternalSymbol struct {
	Symbol string
}

func NewDynValFromString(str string, env *glisp.Glisp) *DynVal {
	sexp, err := env.ParseStream(strings.NewReader(str))
	if err != nil {
		return nil
	}
	return &DynVal{sexp[0], sexp[0].SexpString()}
}

func NewDynValFromSexp(sexp glisp.Sexp) *DynVal {
	return &DynVal{sexp, sexp.SexpString()}
}

func SetClientData(env *glisp.Glisp, cdata *ClientData) error {
	env.AddGlobal("APP_KEY", glisp.SexpStr(cdata.AppKey))
	env.AddGlobal("OS_TYPE", glisp.SexpStr(cdata.OSType))
	env.AddGlobal("OS_VERSION", glisp.SexpStr(cdata.OSVersion))
	env.AddGlobal("APP_VERSION", glisp.SexpStr(cdata.AppVersion))
	env.AddGlobal("IP", glisp.SexpStr(cdata.Ip))
	env.AddGlobal("LANG", glisp.SexpStr(cdata.Lang))
	env.AddGlobal("DEVICE_ID", glisp.SexpStr(cdata.DeviceId))
	return nil
}

func ClearClientData(env *glisp.Glisp) error {
	env.AddGlobal("APP_KEY", glisp.SexpNull)
	env.AddGlobal("OS_TYPE", glisp.SexpNull)
	env.AddGlobal("OS_VERSION", glisp.SexpNull)
	env.AddGlobal("APP_VERSION", glisp.SexpNull)
	env.AddGlobal("IP", glisp.SexpNull)
	env.AddGlobal("LANG", glisp.SexpNull)
	env.AddGlobal("DEVICE_ID", glisp.SexpNull)
	return nil
}

func (dval *DynVal) Execute(env *glisp.Glisp) (glisp.Sexp, error) {
	env.LoadExpressions([]glisp.Sexp{dval.Sexp})
	sexp, err := env.Run()
	if err != nil {
		return glisp.SexpNull, err
	}
	return sexp, nil
}

func EvalDynValToSexp(code string, cdata *ClientData) (glisp.Sexp, error) {
	env := NewGlisp()
	SetClientData(env, cdata)
	dval := NewDynValFromString(code, env)
	return dval.Execute(env)
}

func EvalDynVal(code string, cdata *ClientData) interface{} {
	data, err := EvalDynValToSexp(code, cdata)
	if err != nil {
		return nil
	}
	switch val := data.(type) {
	case glisp.SexpBool:
		return bool(val)
	case glisp.SexpInt:
		return int(val)
	case glisp.SexpFloat:
		return float64(val)
	case glisp.SexpStr:
		return string(val)
	default:
		return data.SexpString()
	}
}

func sexpToSlice(sexp glisp.Sexp) interface{} {
	if sexp == glisp.SexpNull {
		return nil
	}
	switch val := sexp.(type) {
	case glisp.SexpPair:
		return sexpPairToSlice(val)
	case glisp.SexpSymbol:
		return ExternalSymbol{val.Name()}
	case glisp.SexpBool:
		return bool(val)
	case glisp.SexpInt:
		return int(val)
	case glisp.SexpFloat:
		return float64(val)
	case glisp.SexpStr:
		return string(val)
	default:
		return sexp.SexpString()
	}
}

func sexpPairToSlice(pair glisp.SexpPair) []interface{} {
	retv := []interface{}{}

	for {
		switch tail := pair.Tail().(type) {
		case glisp.SexpPair:
			retv = append(retv, sexpToSlice(pair.Head()))
			pair = tail
			continue
		}
		break
	}

	retv = append(retv, sexpToSlice(pair.Head()))
	// TODO fake list when pair.tail is not SexpNull
	return retv
}

func (dval *DynVal) ToSlice() []interface{} {
	return sexpPairToSlice(dval.Sexp.(glisp.SexpPair))
}

func (dval *DynVal) ToJson() (string, error) {
	data, err := json.Marshal(dval.ToSlice())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func sliceToSexpString(data []interface{}) string {
	ret := "("
	for idx, item := range data {
		switch val := item.(type) {
		case bool:
			ret += " "
			ret += strconv.FormatBool(val)
		case int:
			ret += " "
			ret += string(val)
		case float64:
			ret += " "
			ret += strconv.FormatFloat(val, 'f', -1, 64)
		case string:
			ret += " "
			ret += `"`
			ret += string(val)
			ret += `"`
		case map[string]interface{}: // Symbol
			if idx != 0 {
				ret += " "
			}
			ret += string(val["Symbol"].(string))
		case []interface{}: // Sub sexp
			ret += " "
			ret += sliceToSexpString(val)
		}
	}
	ret += ")"
	return ret
}

func JsonToSexpString(json_str string) (string, error) {
	var f []interface{}
	err := json.Unmarshal([]byte(json_str), &f)
	if err != nil {
		return "", err
	}
	data := sliceToSexpString(f)
	return data, nil
}
