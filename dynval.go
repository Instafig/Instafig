package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/zhemao/glisp/interpreter"
)

type glispSymbolValChecker func(val interface{}) error
type glispSymbolType int

const (
	GLISP_SYMBOL_TYPE_APP_VERSION glispSymbolType = iota
	GLISP_SYMBOL_TYPE_OS_VERSION
	GLISP_SYMBOL_TYPE_LANG
	GLISP_SYMBOL_TYPE_OS_TYPE
	GLISP_SYMBOL_TYPE_DEVICE_ID
	GLISP_SYMBOL_TYPE_TIMEZONE
	GLISP_SYMBOL_TYPE_NETWORK
	GLISP_SYMBOL_TYPE_IP
)

type glispSymbolContext struct {
	name      string
	typ       glispSymbolType
	checkFunc glispSymbolValChecker
}

var glispSymbolContexts = []glispSymbolContext{
	{"OS_TYPE", GLISP_SYMBOL_TYPE_OS_TYPE, stringSymbolChecker},
	{"OS_VERSION", GLISP_SYMBOL_TYPE_OS_VERSION, versionSymbolChecker},
	{"APP_VERSION", GLISP_SYMBOL_TYPE_APP_VERSION, versionSymbolChecker},
	{"LANG", GLISP_SYMBOL_TYPE_LANG, stringSymbolChecker},
	{"DEVICE_ID", GLISP_SYMBOL_TYPE_DEVICE_ID, stringSymbolChecker},
	{"TIMEZONE", GLISP_SYMBOL_TYPE_TIMEZONE, stringSymbolChecker},
	{"NETWORK", GLISP_SYMBOL_TYPE_NETWORK, stringSymbolChecker},
	{"IP", GLISP_SYMBOL_TYPE_IP, stringSymbolChecker},
}

func stringSymbolChecker(val interface{}) error {
	switch val.(type) {
	case string:
		return nil
	default:
		return fmt.Errorf("symbol value shoule be string type")
	}
}

func versionSymbolChecker(val interface{}) error {
	switch data := val.(type) {
	case string:
		if _, err := version.NewVersion(data); err != nil {
			return fmt.Errorf("bad version format: [%s] - %s", data, err.Error())
		}
	default:
		return fmt.Errorf("symbol value shoule be string type")
	}

	return nil
}

type glispFuncContext struct {
	name string
	// for positive is accurate arg number, for negative is at least arg numer
	// for example: if argNum is 2, this func needs 2 args; if argNum is -3, this func needs at least 3 args
	argNum int

	// if len(supportSymbols) == 0, do not check
	supportSymbols []glispSymbolType

	multiArgumentsSymbol bool
}

var glispFuncContexts = []glispFuncContext{
	// glisp built-in func
	{"and", -2, nil, true},
	{"or", -2, nil, true},
	{"not", 1, nil, true},

	// version-cmp func
	{"version-cmp",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver>",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver>=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver<",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver<=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},
	{"ver!=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_APP_VERSION, GLISP_SYMBOL_TYPE_OS_VERSION},
		false,
	},

	// str func
	{"str=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str!=",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-empty?",
		1,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-not-empty?",
		1,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-wcmatch?",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-not-wcmatch?",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-contains?",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
	{"str-not-contains?",
		2,
		[]glispSymbolType{GLISP_SYMBOL_TYPE_IP, GLISP_SYMBOL_TYPE_LANG, GLISP_SYMBOL_TYPE_DEVICE_ID, GLISP_SYMBOL_TYPE_NETWORK, GLISP_SYMBOL_TYPE_TIMEZONE, GLISP_SYMBOL_TYPE_OS_TYPE},
		false,
	},
}

var (
	supportedSymbolContexts = map[string]*glispSymbolContext{}

	// all service-defined func must added to supportedFunc
	supportedFuncContexts = map[string]*glispFuncContext{}
)

func init() {
	for _, symbolContext := range glispSymbolContexts {
		s := symbolContext
		supportedSymbolContexts[symbolContext.name] = &s
	}
	for _, funcContext := range glispFuncContexts {
		s := funcContext
		supportedFuncContexts[funcContext.name] = &s
	}
}

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
func plainDataToSexpString(data interface{}, funcContext *glispFuncContext, symbolContext *glispSymbolContext) (string, error) {
	switch data := data.(type) {
	case bool:
		if symbolContext != nil && symbolContext.checkFunc != nil {
			if err := symbolContext.checkFunc(data); err != nil {
				return "", err
			}

		}
		return strconv.FormatBool(data), nil

	case int:
		if symbolContext != nil && symbolContext.checkFunc != nil {
			if err := symbolContext.checkFunc(data); err != nil {
				return "", err
			}

		}
		return string(data), nil

	case float64:
		if symbolContext != nil && symbolContext.checkFunc != nil {
			if err := symbolContext.checkFunc(data); err != nil {
				return "", err
			}

		}
		return strconv.FormatFloat(data, 'f', -1, 64), nil

	case string:
		if symbolContext != nil && symbolContext.checkFunc != nil {
			if err := symbolContext.checkFunc(data); err != nil {
				return "", err
			}

		}
		return `"` + data + `"`, nil // TODO escape double quote

	case map[string]interface{}:
		if val, ok := data["cond-values"]; ok { // cond-values exp
			ret := "(cond-values"
			conds := val.([]interface{})
			for _, cond := range conds {
				ret += " "
				s, err := plainDataToSexpString(cond.(map[string]interface{})["condition"], nil, nil)
				if err != nil {
					return "", err
				}
				ret += s
				ret += " "
				s, err = plainDataToSexpString(cond.(map[string]interface{})["value"], nil, nil)
				if err != nil {
					return "", err
				}
				ret += s
			}
			if dft, ok := data["default-value"]; ok {
				ret += " "
				s, err := plainDataToSexpString(dft, nil, nil)
				if err != nil {
					return "", nil
				}
				ret += s
			}
			ret += ")"
			return ret, nil
		}

		if val, ok := data["func"]; ok { // cond-values style function call
			// shadow super symbolContext
			symbolContext = nil
			funcContext = supportedFuncContexts[val.(string)]
			if funcContext == nil {
				return "", fmt.Errorf("unknown func: " + val.(string))
			}

			ret := "(" + val.(string)
			args := data["arguments"].([]interface{})

			// we are in arg list of a func
			// 1. check arg number
			switch {
			case len(args) == 0:
				return "", fmt.Errorf("must have a symbol in func <%s> arg list", funcContext.name)
			case funcContext.argNum >= 0 && len(args) != funcContext.argNum:
				return "", fmt.Errorf("func <%s> must have %d args", funcContext.name, funcContext.argNum)
			case funcContext.argNum < 0 && len(args) < -funcContext.argNum:
				return "", fmt.Errorf("func <%s> must have at least %d args", funcContext.name, funcContext.argNum)
			}

			if !funcContext.multiArgumentsSymbol {
				// 2. check symbol validity
				var symbol map[string]interface{}
				var ok bool
				if symbol, ok = args[0].(map[string]interface{}); !ok {
					return "", fmt.Errorf("1st element of func <%s> arg list must be symbol", funcContext.name)
				}
				symbolContext = supportedSymbolContexts[symbol["symbol"].(string)]
				if symbolContext == nil {
					return "", fmt.Errorf("unsupported symbol: %s", symbol["symbol"].(string))
				}
				if len(funcContext.supportSymbols) > 0 {
					ok = false
					for _, _symbolContext := range funcContext.supportSymbols {
						if symbolContext.typ == _symbolContext {
							ok = true
							break
						}
					}
					if !ok {
						return "", fmt.Errorf("symbol <%s> is not supported in func <%s>", symbolContext.name, funcContext.name)
					}
				}
			}

			for _, argval := range args {
				ret += " "
				s, err := plainDataToSexpString(argval, funcContext, symbolContext)
				if err != nil {
					return "", err
				}
				ret += s
			}
			ret += ")"
			return ret, nil
		}

		if val, ok := data["symbol"]; ok { // Symbol
			return string(val.(string)), nil
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
			s, err := plainDataToSexpString(val, funcContext, symbolContext)
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

	return plainDataToSexpString(f, nil, nil)
}

func CheckJsonString(j string) error {
	sexp, err := JsonToSexpString(j)
	if err != nil {
		return err
	}

	_, err = EvalDynVal(NewDynValFromSexpStringDefault(sexp), &ClientData{})

	return err
}
