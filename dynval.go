package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/zhemao/glisp/interpreter"
)

type DynVal struct {
	Sexp     glisp.Sexp
	Sexp_str string
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

func  NewDynValFromStringDefault(sexp string) *DynVal {
	env := NewGlisp()
	SetClientData(env, &ClientData{})
	return NewDynValFromString(sexp, env)
}

// Eval
func (dval *DynVal) Execute(env *glisp.Glisp) (glisp.Sexp, error) {
	env.LoadExpressions([]glisp.Sexp{dval.Sexp})
	sexp, err := env.Run()
	if err != nil {
		return glisp.SexpNull, err
	}
	return sexp, nil
}

func EvalDynValToSexp(code *DynVal, cdata *ClientData) (glisp.Sexp, error) {
	env := NewGlisp()
	SetClientData(env, cdata)
	//dval := NewDynValFromString(code, env)
	return code.Execute(env)
}

func EvalDynVal(code *DynVal, cdata *ClientData) interface{} {
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

func EvalDynValFromExpString(code string, cdata *ClientData) interface{} {
	env := NewGlisp()
	SetClientData(env, cdata)
	dval := NewDynValFromString(code, env)
	data, err := dval.Execute(env)

	//env := NewGlisp()
	//SetClientData(env, cdata)
	//dyval := NewDynValFromString(code, env)
	//data, err := dyval.Execute(env)
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

// Serialize to JSON

// 1. for normal style
func sexpToPlainData(sexp glisp.Sexp) interface{} {
	if sexp == glisp.SexpNull {
		return nil
	}
	switch val := sexp.(type) {
	case glisp.SexpPair:
		return sexpPairToPlainData(val)
	case glisp.SexpSymbol:
		ret := make(map[string]string)
		ret["symbol"] = val.Name()
		return ret
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

func sexpPairToPlainData(pair glisp.SexpPair) interface{} {
	retv := []interface{}{}

	if h, ok := pair.Head().(glisp.SexpSymbol); ok {
		if h.Name() == "cond-values" {
			return condValuesToPlainData(pair, false)
		}
	}

	for {
		switch tail := pair.Tail().(type) {
		case glisp.SexpPair:
			retv = append(retv, sexpToPlainData(pair.Head()))
			pair = tail
			continue
		}
		break
	}

	retv = append(retv, sexpToPlainData(pair.Head()))
	// TODO fake list when pair.tail is not SexpNull
	return retv
}

// 2. for cond-values style

func condValuesBodyToPlainData(body glisp.SexpPair) ([]interface{}, interface{}) {
	var default_value interface{}
	var conds []interface{}
	for {
		cond := body.Head()
		if tail, ok := body.Tail().(glisp.SexpPair); ok {
			body = tail
		} else {
			default_value = condValuesToPlainData(cond, false)
			break
		}
		value := body.Head()
		cond_item := make(map[string]interface{})
		cond_item["condition"] = condValuesToPlainData(cond, false)
		cond_item["value"] = condValuesToPlainData(value, false)
		conds = append(conds, cond_item)
		if tail, ok := body.Tail().(glisp.SexpPair); ok {
			body = tail
		} else {
			default_value = nil
			break
		}
	}
	return conds, default_value
}

func condValuesToPlainData(sexp glisp.Sexp, issub bool) interface{} {
	switch expv := sexp.(type) {
	case glisp.SexpPair:
		switch val := expv.Head().(type) {
		case glisp.SexpSymbol:
			if issub {
				retv := []interface{}{}
				retv = append(retv, sexpToPlainData(val))
				if rest, ok := (expv.Tail()).(glisp.SexpPair); ok {
					for {
						switch tail := rest.Tail().(type) {
						case glisp.SexpPair:
							retv = append(retv, condValuesToPlainData(rest.Head(), false))
							rest = tail
							continue
						}
						break
					}
					retv = append(retv, condValuesToPlainData(rest.Head(), false))
				}
				return retv
			}
			if val.Name() == "cond-values" {
				ret := make(map[string]interface{})
				if tail, ok := expv.Tail().(glisp.SexpPair); ok {
					ret["cond-values"], ret["default-value"] = condValuesBodyToPlainData(tail)
				} else {
					ret["cond-values"], ret["default-value"] = []interface{}{}, condValuesToPlainData(tail, false)
				}
				return ret
			} else {
				ret := make(map[string]interface{})
				ret["func"] = val.Name()
				ret["arguments"] = condValuesToPlainData(expv.Tail(), true)
				return ret
			}
		default:
			retv := []interface{}{}
			for {
				switch tail := expv.Tail().(type) {
				case glisp.SexpPair:
					retv = append(retv, condValuesToPlainData(expv.Head(), false))
					expv = tail
					continue
				}
				break
			}
			retv = append(retv, condValuesToPlainData(expv.Head(), false))
			return retv
		}
	default:
		return sexpToPlainData(sexp)
	}
}

// 3. API
func (dval *DynVal) ToPlainData() interface{} {
	return sexpToPlainData(dval.Sexp)
}

func (dval *DynVal) ToJson() (string, error) {
	data, err := json.Marshal(dval.ToPlainData())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Unserialize from JSON to Sexp
func plainDataToSexpString(data interface{}) string {
	switch data := data.(type) {
	case bool:
		return strconv.FormatBool(data)
	case int:
		return string(data)
	case float64:
		return strconv.FormatFloat(data, 'f', -1, 64)
	case string:
		return `"` + data + `"` // TODO escape double quote
	case map[string]interface{}:
		if val, ok := data["symbol"]; ok { // Symbol
			return string(val.(string))
		}
		if val, ok := data["func"]; ok { // cond-values style function call
			ret := "(" + val.(string)
			args := data["arguments"].([]interface{})
			for _, argval := range args {
				ret += " "
				ret += plainDataToSexpString(argval)
			}
			ret += ")"
			return ret
		}
		if val, ok := data["cond-values"]; ok { // cond-values exp
			ret := "(cond-values"
			conds := val.([]interface{})
			for _, cond := range conds {
				ret += " "
				ret += plainDataToSexpString(cond.(map[string]interface{})["condition"])
				ret += " "
				ret += plainDataToSexpString(cond.(map[string]interface{})["value"])
			}
			if dft, ok := data["default-value"]; ok {
				ret += " "
				ret += plainDataToSexpString(dft)
			}
			ret += ")"
			return ret
		}

	case []interface{}: // Sub sexp
		ret := "("
		for idx, val := range data {
			if idx != 0 {
				ret += " "
			}
			ret += plainDataToSexpString(val)
		}
		ret += ")"
		return ret
	}
	return "()"
}

func JsonToSexpString(json_str string) (string, error) {
	var f interface{}
	err := json.Unmarshal([]byte(json_str), &f)
	if err != nil {
		return "", err
	}
	data := plainDataToSexpString(f)
	return data, nil
}
