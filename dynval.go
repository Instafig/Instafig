package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zhemao/glisp/interpreter"
)

var (
	supportedSymbol = map[string]bool{
		"APP_KEY":     true,
		"OS_TYPE":     true,
		"OS_VERSION":  true,
		"APP_VERSION": true,
		"LANG":        true,
		"DEVICE_ID":   true,
		"TIMEZONE":    true,
		"NETWORK":     true,
	}

	// all service-defined func must added to supportedFunc
	supportedFunc = map[string]bool{
		// glisp built-in func
		"and": true,
		"or":  true,
		// version func
		"version-cmp": true,
		"ver=":        true,
		"ver>":        true,
		"ver>=":       true,
		"ver<":        true,
		"ver<=":       true,
		"ver!=":       true,
		// str func
		"str=":              true,
		"str!=":             true,
		"str-empty?":        true,
		"str-wcmatch?":      true,
		"str-contains?":     true,
		"str-not-empty?":    true,
		"str-not-contains?": true,
		"str-not-wcmatch?":  true,
	}
)

type DynVal struct {
	Sexp    glisp.Sexp
	SexpStr string
}

func NewDynValFromString(str string, env *glisp.Glisp) *DynVal {
	sexp, err := env.ParseStream(strings.NewReader(str))
	if err != nil {
		return nil
	}
	return &DynVal{sexp[0], sexp[0].SexpString()}
}

func SetClientData(env *glisp.Glisp, cdata *ClientData) error {
	env.AddGlobal("APP_KEY", glisp.SexpStr(cdata.AppKey))
	env.AddGlobal("OS_TYPE", glisp.SexpStr(cdata.OSType))
	env.AddGlobal("OS_VERSION", glisp.SexpStr(cdata.OSVersion))
	env.AddGlobal("APP_VERSION", glisp.SexpStr(cdata.AppVersion))
	env.AddGlobal("IP", glisp.SexpStr(cdata.Ip))
	env.AddGlobal("LANG", glisp.SexpStr(cdata.Lang))
	env.AddGlobal("DEVICE_ID", glisp.SexpStr(cdata.DeviceId))
	env.AddGlobal("TIMEZONE", glisp.SexpStr(cdata.TimeZone))
	env.AddGlobal("NETWORK", glisp.SexpStr(cdata.NetWork))
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
	env.AddGlobal("TIMEZONE", glisp.SexpNull)
	env.AddGlobal("NETWORK", glisp.SexpNull)
	return nil
}

func NewDynValFromSexpStringDefault(sexp string) *DynVal {
	env := getGLispEnv()
	defer putGLispEnv(env)

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
	env := getGLispEnv()
	defer putGLispEnv(env)

	SetClientData(env, cdata)
	//dval := NewDynValFromString(code, env)
	return code.Execute(env)
}

func EvalDynVal(code *DynVal, cdata *ClientData) (interface{}, error) {
	data, err := EvalDynValToSexp(code, cdata)
	if err != nil {
		return nil, err
	}
	switch val := data.(type) {
	case glisp.SexpBool:
		return bool(val), nil
	case glisp.SexpInt:
		return int(val), nil
	case glisp.SexpFloat:
		return float64(val), nil
	case glisp.SexpStr:
		return string(val), nil
	default:
		return data.SexpString(), nil
	}
}

func EvalDynValNoErr(code *DynVal, cdata *ClientData) interface{} {
	i, _ := EvalDynVal(code, cdata)
	return i
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
func plainDataToSexpString(data interface{}) (string, error) {
	switch data := data.(type) {
	case bool:
		return strconv.FormatBool(data), nil

	case int:
		return string(data), nil

	case float64:
		return strconv.FormatFloat(data, 'f', -1, 64), nil

	case string:
		return `"` + data + `"`, nil // TODO escape double quote

	case map[string]interface{}:
		if val, ok := data["symbol"]; ok { // Symbol
			if !supportedSymbol[val.(string)] {
				return "", fmt.Errorf("unknown symbol: " + val.(string))
			}
			return string(val.(string)), nil
		}
		if val, ok := data["func"]; ok { // cond-values style function call
			if !supportedFunc[val.(string)] {
				return "", fmt.Errorf("unknown func: " + val.(string))
			}
			ret := "(" + val.(string)
			args := data["arguments"].([]interface{})
			for _, argval := range args {
				ret += " "
				s, err := plainDataToSexpString(argval)
				if err != nil {
					return "", err
				}
				ret += s
			}
			ret += ")"
			return ret, nil
		}
		if val, ok := data["cond-values"]; ok { // cond-values exp
			ret := "(cond-values"
			conds := val.([]interface{})
			for _, cond := range conds {
				ret += " "
				s, err := plainDataToSexpString(cond.(map[string]interface{})["condition"])
				if err != nil {
					return "", err
				}
				ret += s
				ret += " "
				s, err = plainDataToSexpString(cond.(map[string]interface{})["value"])
				if err != nil {
					return "", err
				}
				ret += s
			}
			if dft, ok := data["default-value"]; ok {
				ret += " "
				s, err := plainDataToSexpString(dft)
				if err != nil {
					return "", nil
				}
				ret += s
			}
			ret += ")"
			return ret, nil
		}

		for k, _ := range data {
			return "", fmt.Errorf("unknown symbol: " + k)
		}

	case []interface{}: // Sub sexp
		ret := "("
		for idx, val := range data {
			if idx != 0 {
				ret += " "
			}
			s, err := plainDataToSexpString(val)
			if err != nil {
				return "", err
			}
			ret += s
		}
		ret += ")"
		return ret, nil
	}

	return "()", nil
}

func JsonToSexpString(json_str string) (string, error) {
	var f interface{}
	err := json.Unmarshal([]byte(json_str), &f)
	if err != nil {
		return "", err
	}

	return plainDataToSexpString(f)
}

func CheckJsonString(j string) error {
	sexp, err := JsonToSexpString(j)
	if err != nil {
		return err
	}

	_, err = EvalDynVal(NewDynValFromSexpStringDefault(sexp), &ClientData{})

	return err
}
